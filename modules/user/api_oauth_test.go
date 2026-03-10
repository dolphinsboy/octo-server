package user

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// parseAccessToken 模拟从 OAuth 响应中解析 access_token 的逻辑
// 这是 requestGiteeAccessToken 和 requestGithubAccessToken 中的核心解析逻辑
func parseAccessToken(result map[string]interface{}) (string, error) {
	accessToken := ""
	if result["access_token"] != nil {
		if token, ok := result["access_token"].(string); ok {
			accessToken = token
		} else {
			return "", errors.New("access_token 类型错误")
		}
	}
	return accessToken, nil
}

func TestParseAccessToken_ValidString(t *testing.T) {
	// 正常情况：access_token 是字符串
	result := map[string]interface{}{
		"access_token": "gho_abc123def456",
		"token_type":   "bearer",
	}
	token, err := parseAccessToken(result)
	assert.NoError(t, err)
	assert.Equal(t, "gho_abc123def456", token)
}

func TestParseAccessToken_NilValue(t *testing.T) {
	// access_token 为 nil
	result := map[string]interface{}{
		"error": "bad_verification_code",
	}
	token, err := parseAccessToken(result)
	assert.NoError(t, err)
	assert.Equal(t, "", token)
}

func TestParseAccessToken_InvalidTypeInt(t *testing.T) {
	// access_token 是 int 类型（异常响应）
	result := map[string]interface{}{
		"access_token": 12345,
	}
	token, err := parseAccessToken(result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access_token 类型错误")
	assert.Equal(t, "", token)
}

func TestParseAccessToken_InvalidTypeFloat(t *testing.T) {
	// access_token 是 float 类型
	result := map[string]interface{}{
		"access_token": 123.456,
	}
	token, err := parseAccessToken(result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access_token 类型错误")
	assert.Equal(t, "", token)
}

func TestParseAccessToken_InvalidTypeMap(t *testing.T) {
	// access_token 是 map 类型
	result := map[string]interface{}{
		"access_token": map[string]interface{}{"nested": "value"},
	}
	token, err := parseAccessToken(result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access_token 类型错误")
	assert.Equal(t, "", token)
}

func TestParseAccessToken_InvalidTypeBool(t *testing.T) {
	// access_token 是 bool 类型
	result := map[string]interface{}{
		"access_token": true,
	}
	token, err := parseAccessToken(result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access_token 类型错误")
	assert.Equal(t, "", token)
}

func TestParseAccessToken_EmptyString(t *testing.T) {
	// access_token 是空字符串
	result := map[string]interface{}{
		"access_token": "",
	}
	token, err := parseAccessToken(result)
	assert.NoError(t, err)
	assert.Equal(t, "", token)
}
