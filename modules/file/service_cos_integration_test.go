package file_test

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Mininglamp-OSS/octo-lib/config"
	"github.com/Mininglamp-OSS/octo-lib/testutil"
	"github.com/Mininglamp-OSS/octo-server/modules/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCOSPresignedURLs_SignAgainstPublicEndpoint mirrors the MinIO-side
// integration test (see service_minio_integration_test.go) for COS.
//
// PR#50 R6 shipped a `presigned.Host = bucketURL.Host` mutation AFTER
// signing — same hazard MinIO closed at R3+. SigV4 covers `host` in
// the signed headers, so any post-sign host change produces 403
// SignatureDoesNotMatch from the COS gateway on every browser PUT/GET.
//
// The R7 fix builds a public-facing minio client whose endpoint is
// derived from `cosConfig.BucketURL` (parent domain after stripping
// the documented `<bucket>.` subdomain), and signs against that
// client directly. Reading the resulting URL host back out and
// confirming it matches BucketURL is equivalent to confirming the
// signature is valid for that host: if the URL host disagreed with
// the host actually signed, the URL would not authenticate at the
// COS gateway.
//
// The test uses fake credentials and never makes a network call —
// PresignHeader / PresignedGetObject are pure URL signing.
func TestCOSPresignedURLs_SignAgainstPublicEndpoint(t *testing.T) {
	cfg := config.New()
	cfg.Test = true
	cfg.COS.SecretID = "test-secret-id"
	cfg.COS.SecretKey = "test-secret-key-1234567890"
	cfg.COS.Bucket = "my-bucket-12345678"
	cfg.COS.Region = "ap-beijing"
	cfg.COS.BucketURL = "https://my-bucket-12345678.cos.example.com"

	ctx := testutil.NewTestContext(cfg)
	svc := file.NewServiceCOS(ctx)

	t.Run("PUT URL signed against public host (no rewriting)", func(t *testing.T) {
		uploadURL, downloadURL, err := svc.PresignedPutURL(
			"chat/2026/05/abc.jpg", "image/jpeg", "", 12345, 5*time.Minute,
		)
		require.NoError(t, err)
		require.NotEmpty(t, uploadURL)
		require.NotEmpty(t, downloadURL)

		u, err := url.Parse(uploadURL)
		require.NoError(t, err)

		// Host check: BucketURL host should match exactly. The minio
		// SDK virtual-hosts `<bucket>.<parent>` — with parent
		// `cos.example.com` and bucket `my-bucket-12345678`, the
		// reconstructed host equals BucketURL host.
		assert.Equal(t, "my-bucket-12345678.cos.example.com", u.Host,
			"presigned PUT URL must be served from the BucketURL host, got %s", u.Host)
		assert.Equal(t, "https", u.Scheme,
			"presigned PUT URL must inherit scheme from BucketURL")

		// SigV4 shape: `host` and `content-length` MUST appear in
		// the signed headers. Because the signing client was
		// constructed against BucketURL's parent domain, the host
		// covered by the signature is the URL's own host. Any
		// post-sign host change would break that invariant.
		// `content-length` is the P0 size-bypass guard landed in R6.
		q := u.Query()
		assert.NotEmpty(t, q.Get("X-Amz-Signature"),
			"presigned PUT URL must carry a SigV4 signature")
		signedHeaders := q.Get("X-Amz-SignedHeaders")
		assert.Contains(t, signedHeaders, "host",
			"presigned PUT URL must include `host` in its signed headers (got %q)", signedHeaders)
		assert.Contains(t, signedHeaders, "content-length",
			"presigned PUT URL must include `content-length` in its signed headers so the COS gateway can enforce the upload size cap (got %q)", signedHeaders)
	})

	t.Run("GET URL signed against public host (no rewriting)", func(t *testing.T) {
		raw, err := svc.PresignedGetURL("chat/2026/05/abc.jpg", "report.jpg", "attachment", 5*time.Minute)
		require.NoError(t, err)
		require.NotEmpty(t, raw)

		u, err := url.Parse(raw)
		require.NoError(t, err)

		assert.Equal(t, "my-bucket-12345678.cos.example.com", u.Host,
			"presigned GET URL must be served from the BucketURL host, got %s", u.Host)
		assert.Equal(t, "https", u.Scheme,
			"presigned GET URL must inherit scheme from BucketURL")

		q := u.Query()
		assert.NotEmpty(t, q.Get("X-Amz-Signature"),
			"presigned GET URL must carry a SigV4 signature")
		assert.NotEmpty(t, q.Get("X-Amz-Credential"),
			"presigned GET URL must carry the SigV4 credential scope")
		signedHeaders := q.Get("X-Amz-SignedHeaders")
		assert.Contains(t, signedHeaders, "host",
			"presigned GET URL must include `host` in its signed headers (got %q)", signedHeaders)

		assert.True(t,
			strings.Contains(u.Path, "/chat/") && strings.HasSuffix(u.Path, "/abc.jpg"),
			"object path should be reflected in the signed URL, got %s", u.Path)

		disposition := q.Get("response-content-disposition")
		assert.Contains(t, disposition, "attachment",
			"response-content-disposition should preserve the requested disposition")
		assert.Contains(t, disposition, "report.jpg",
			"response-content-disposition should carry the requested filename")
	})
}

// TestCOSPresignedURLs_DefaultEndpointWhenBucketURLEmpty pins the
// fallback contract: when `cosConfig.BucketURL` is empty, presigned
// URLs are signed against the SDK's canonical endpoint
// `<bucket>.cos.<region>.myqcloud.com`. This is the COS "no custom
// domain" deployment shape — the canonical hostname is reachable
// from the browser without any operator-side DNS work.
func TestCOSPresignedURLs_DefaultEndpointWhenBucketURLEmpty(t *testing.T) {
	cfg := config.New()
	cfg.Test = true
	cfg.COS.SecretID = "test-secret-id"
	cfg.COS.SecretKey = "test-secret-key-1234567890"
	cfg.COS.Bucket = "my-bucket-12345678"
	cfg.COS.Region = "ap-beijing"
	cfg.COS.BucketURL = "" // fallback path

	svc := file.NewServiceCOS(testutil.NewTestContext(cfg))

	uploadURL, _, err := svc.PresignedPutURL(
		"chat/2026/05/abc.jpg", "image/jpeg", "", 12345, time.Minute,
	)
	require.NoError(t, err)

	u, err := url.Parse(uploadURL)
	require.NoError(t, err)
	assert.Equal(t, "my-bucket-12345678.cos.ap-beijing.myqcloud.com", u.Host,
		"with BucketURL empty, presigned URL must use canonical COS host")
	assert.Equal(t, "https", u.Scheme,
		"COS canonical endpoint must be HTTPS")
}

// TestCOSPresignedURLs_WithPrefix pins the env-prefix routing: when
// `cosConfig.Prefix` is set (multi-env shared bucket), the prefix
// is prepended to the object key BEFORE signing, so the signed URL
// resolves to the prefixed object on the COS server. This is the
// behaviour `withPrefix` provided in R6 and the R7 host fix must
// not regress.
func TestCOSPresignedURLs_WithPrefix(t *testing.T) {
	cfg := config.New()
	cfg.Test = true
	cfg.COS.SecretID = "test-secret-id"
	cfg.COS.SecretKey = "test-secret-key-1234567890"
	cfg.COS.Bucket = "my-bucket-12345678"
	cfg.COS.Region = "ap-beijing"
	cfg.COS.BucketURL = "https://my-bucket-12345678.cos.example.com"
	cfg.COS.Prefix = "env-test-prefix"

	svc := file.NewServiceCOS(testutil.NewTestContext(cfg))

	uploadURL, _, err := svc.PresignedPutURL(
		"chat/2026/05/abc.jpg", "image/jpeg", "", 12345, time.Minute,
	)
	require.NoError(t, err)

	u, err := url.Parse(uploadURL)
	require.NoError(t, err)
	assert.Equal(t, "my-bucket-12345678.cos.example.com", u.Host,
		"prefix routing must not perturb the BucketURL host")
	assert.Contains(t, u.Path, "/env-test-prefix/chat/2026/05/abc.jpg",
		"signed URL path must include the env prefix, got %s", u.Path)
}

// TestCOSPresignedURLs_HTTPScheme pins that an `http://` BucketURL is
// honoured (non-TLS local emulators or test setups). Going via the
// SDK's `Secure: false` toggle means the signature is computed for
// the http variant — flipping to https post-sign would invalidate it.
func TestCOSPresignedURLs_HTTPScheme(t *testing.T) {
	cfg := config.New()
	cfg.Test = true
	cfg.COS.SecretID = "test-secret-id"
	cfg.COS.SecretKey = "test-secret-key-1234567890"
	cfg.COS.Bucket = "my-bucket-12345678"
	cfg.COS.Region = "ap-beijing"
	cfg.COS.BucketURL = "http://my-bucket-12345678.cos.local"

	svc := file.NewServiceCOS(testutil.NewTestContext(cfg))

	uploadURL, _, err := svc.PresignedPutURL(
		"chat/2026/05/abc.jpg", "image/jpeg", "", 12345, time.Minute,
	)
	require.NoError(t, err)

	u, err := url.Parse(uploadURL)
	require.NoError(t, err)
	assert.Equal(t, "http", u.Scheme, "http BucketURL must produce http presigned URL")
	assert.Equal(t, "my-bucket-12345678.cos.local", u.Host)
}

// TestServiceCOS_PresignedPutURL_PathStyleCDN pins the YUJ-846 hotfix:
// when `cosConfig.BucketURL` is a custom CDN / accelerator domain that
// does NOT carry a `<bucket>.` subdomain (e.g.
// `https://cdn.example.com`), presigned URLs must be served from that
// host *as-is*, with the bucket placed in the URL path
// (`<host>/<bucket>/<key>`), not virtual-hosted onto a phantom
// subdomain.
//
// Pre-fix behaviour (broken in PR#50 R8 / e8b03a9):
//   - publicEndpoint returned the BucketURL host with no `<bucket>.`
//     prefix to strip, so it kept `cdn.example.com` and reported it
//     as if it were the parent of a virtual-hosted bucket
//   - newPublicClient hardcoded BucketLookupDNS
//   - the SDK then virtual-hosted: `<bucket>.cdn.example.com`, a
//     hostname that does not exist in DNS
//   - browser PUT → `net::ERR_NAME_NOT_RESOLVED`, all uploads broken
//
// Post-fix behaviour (this hotfix):
//   - publicEndpoint detects the missing `<bucket>.` prefix and
//     returns `BucketLookupPath`
//   - newPublicClient threads the lookup style through to minio.New
//   - the SDK signs against `cdn.example.com` exactly and emits
//     `https://cdn.example.com/<bucket>/<key>` — the host the browser
//     actually resolves
//
// This test mirrors the production repro from im-test.deepminer.com.cn
// (BucketURL=`https://cdn.deepminer.com.cn`, bucket=`im-data-...`).
// It uses fake credentials and never makes a network call —
// PresignHeader / PresignedGetObject are pure URL signing.
func TestServiceCOS_PresignedPutURL_PathStyleCDN(t *testing.T) {
	cfg := config.New()
	cfg.Test = true
	cfg.COS.SecretID = "test-secret-id"
	cfg.COS.SecretKey = "test-secret-key-1234567890"
	cfg.COS.Bucket = "im-data-1255521909"
	cfg.COS.Region = "ap-beijing"
	// Path-style CDN: host has NO `<bucket>.` subdomain.
	cfg.COS.BucketURL = "https://cdn.example.com"

	svc := file.NewServiceCOS(testutil.NewTestContext(cfg))

	t.Run("PUT URL is path-style on the CDN host", func(t *testing.T) {
		uploadURL, _, err := svc.PresignedPutURL(
			"chat/2026/05/abc.jpg", "image/jpeg", "", 12345, 5*time.Minute,
		)
		require.NoError(t, err)
		require.NotEmpty(t, uploadURL)

		u, err := url.Parse(uploadURL)
		require.NoError(t, err)

		// Host MUST be the CDN host as-is — NOT
		// `<bucket>.cdn.example.com`. Pre-fix this assertion failed
		// because BucketLookupDNS produced the phantom subdomain.
		assert.Equal(t, "cdn.example.com", u.Host,
			"path-style BucketURL must keep the CDN host verbatim, not virtual-host the bucket onto it; got %s", u.Host)
		assert.NotContains(t, u.Host, "im-data-1255521909",
			"path-style BucketURL must NOT prepend the bucket as a subdomain; got %s", u.Host)

		// Path MUST start with `/<bucket>/` — that's the path-style
		// addressing the CDN expects.
		assert.True(t, strings.HasPrefix(u.Path, "/im-data-1255521909/"),
			"path-style URL must place bucket in the path; got path=%s", u.Path)
		assert.True(t, strings.HasSuffix(u.Path, "/chat/2026/05/abc.jpg"),
			"object key must be reflected in the signed URL path; got path=%s", u.Path)

		assert.Equal(t, "https", u.Scheme,
			"presigned PUT URL must inherit scheme from BucketURL")

		// SigV4 shape: `host` and `content-length` MUST appear in the
		// signed headers. Because the signing client was constructed
		// against the CDN host with BucketLookupPath, the host
		// covered by the signature is the URL's own host
		// (`cdn.example.com`). A reviewer reading the URL back can
		// confirm signature validity by host equality alone.
		q := u.Query()
		assert.NotEmpty(t, q.Get("X-Amz-Signature"),
			"presigned PUT URL must carry a SigV4 signature")
		signedHeaders := q.Get("X-Amz-SignedHeaders")
		assert.Contains(t, signedHeaders, "host",
			"presigned PUT URL must include `host` in its signed headers (got %q)", signedHeaders)
		assert.Contains(t, signedHeaders, "content-length",
			"presigned PUT URL must include `content-length` in its signed headers (got %q)", signedHeaders)
	})

	t.Run("GET URL is path-style on the CDN host", func(t *testing.T) {
		raw, err := svc.PresignedGetURL(
			"chat/2026/05/abc.jpg", "report.jpg", "attachment", 5*time.Minute,
		)
		require.NoError(t, err)
		require.NotEmpty(t, raw)

		u, err := url.Parse(raw)
		require.NoError(t, err)

		assert.Equal(t, "cdn.example.com", u.Host,
			"path-style BucketURL must keep the CDN host verbatim for GET as well; got %s", u.Host)
		assert.True(t, strings.HasPrefix(u.Path, "/im-data-1255521909/"),
			"path-style GET URL must place bucket in the path; got path=%s", u.Path)
		assert.True(t, strings.HasSuffix(u.Path, "/chat/2026/05/abc.jpg"),
			"object key must be reflected in the signed GET URL; got path=%s", u.Path)

		signedHeaders := u.Query().Get("X-Amz-SignedHeaders")
		assert.Contains(t, signedHeaders, "host",
			"presigned GET URL must include `host` in its signed headers (got %q)", signedHeaders)
	})
}

// TestServiceCOS_PresignedPutURL_PathStyleCDN_WithPrefix pins that the
// env-prefix routing keeps working under path-style addressing — the
// prefix is prepended to the object key before signing, and the bucket
// still lands in the URL path (NOT folded into the host).
func TestServiceCOS_PresignedPutURL_PathStyleCDN_WithPrefix(t *testing.T) {
	cfg := config.New()
	cfg.Test = true
	cfg.COS.SecretID = "test-secret-id"
	cfg.COS.SecretKey = "test-secret-key-1234567890"
	cfg.COS.Bucket = "im-data-1255521909"
	cfg.COS.Region = "ap-beijing"
	cfg.COS.BucketURL = "https://cdn.example.com"
	cfg.COS.Prefix = "im-test"

	svc := file.NewServiceCOS(testutil.NewTestContext(cfg))

	uploadURL, _, err := svc.PresignedPutURL(
		"chat/2026/05/abc.jpg", "image/jpeg", "", 12345, time.Minute,
	)
	require.NoError(t, err)

	u, err := url.Parse(uploadURL)
	require.NoError(t, err)
	assert.Equal(t, "cdn.example.com", u.Host,
		"prefix routing under path-style must not perturb the CDN host; got %s", u.Host)
	assert.True(t, strings.HasPrefix(u.Path, "/im-data-1255521909/im-test/chat/2026/05/abc.jpg"),
		"path-style URL must include `/<bucket>/<prefix>/<key>`; got path=%s", u.Path)
}
