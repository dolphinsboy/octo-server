package space

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedUserVerification 往 user_verification 写一条实名记录（issue #344 display-name 兜底链）。
func seedUserVerification(t *testing.T, uid, realName string) {
	t.Helper()
	_, err := testCtx.DB().InsertBySql(
		"INSERT INTO user_verification (user_id, real_name, source, source_sub, verified_at) VALUES (?, ?, 'aegis', ?, NOW()) "+
			"ON DUPLICATE KEY UPDATE real_name=VALUES(real_name)",
		uid, realName, "sub-"+uid,
	).Exec()
	require.NoError(t, err)
}

// findMemberDetail 在 queryMembers 结果里按 uid 找成员。
func findMemberDetail(list []*MemberDetailModel, uid string) (*MemberDetailModel, bool) {
	for _, m := range list {
		if m.UID == uid {
			return m, true
		}
	}
	return nil, false
}

// TestQueryMembersDisplayNameFallback 验证 member 列表的 display-name 兜底链（issue #344）：
//
//	user.name → user_verification.real_name（已实名用户）→ "User-" + left(uid,6) 占位符
//
// 修复前空间成员若 user.name 为空（OIDC IdP 未下发 name claim）会返回空 name 字段。
func TestQueryMembersDisplayNameFallback(t *testing.T) {
	_, _, err := setup(t)
	require.NoError(t, err)

	spaceId := "sp-name-fallback"
	owner := "owner-name-fallback"
	seedMemberSearchSpace(t, spaceId, owner)

	const (
		namedUID       = "uid-named-001"
		verifiedUID    = "uid-verified-002"
		placeholderUID = "uidplaceholder003"
	)

	// 1. user.name 非空 → 直接用 user.name
	seedMemberSearchUser(t, namedUID, "Alice", "alice", "", "")
	seedMemberSearchMember(t, spaceId, namedUID, 0, 1)

	// 2. user.name 空但 user_verification.real_name 有值 → 用 real_name
	seedMemberSearchUser(t, verifiedUID, "", "bob", "", "")
	seedUserVerification(t, verifiedUID, "Bob Real")
	seedMemberSearchMember(t, spaceId, verifiedUID, 0, 1)

	// 3. user.name 空且无实名记录 → "User-" + left(uid,6) 占位符
	seedMemberSearchUser(t, placeholderUID, "", "carol", "", "")
	seedMemberSearchMember(t, spaceId, placeholderUID, 0, 1)

	members, err := testSpaceDB.queryMembers(spaceId, owner, 1, 50)
	require.NoError(t, err)

	named, ok := findMemberDetail(members, namedUID)
	require.True(t, ok, "named member should be present")
	assert.Equal(t, "Alice", named.Name)

	verified, ok := findMemberDetail(members, verifiedUID)
	require.True(t, ok, "verified member should be present")
	assert.Equal(t, "Bob Real", verified.Name)

	placeholder, ok := findMemberDetail(members, placeholderUID)
	require.True(t, ok, "placeholder member should be present")
	assert.Equal(t, "User-"+placeholderUID[:6], placeholder.Name)
	assert.NotEmpty(t, placeholder.Name, "name must never be empty (issue #344)")
}

// TestSearchMembersDisplayNameFallback 验证成员搜索（空 keyword 返回全部）走同一条兜底链。
func TestSearchMembersDisplayNameFallback(t *testing.T) {
	_, _, err := setup(t)
	require.NoError(t, err)

	spaceId := "sp-search-name-fallback"
	owner := "owner-search-fallback"
	seedMemberSearchSpace(t, spaceId, owner)

	const (
		namedUID       = "uid-s-named-001"
		verifiedUID    = "uid-s-verified-002"
		placeholderUID = "uidsplaceholder03"
	)

	seedMemberSearchUser(t, namedUID, "Alice", "s-alice", "", "")
	seedMemberSearchMember(t, spaceId, namedUID, 0, 1)

	seedMemberSearchUser(t, verifiedUID, "", "s-bob", "", "")
	seedUserVerification(t, verifiedUID, "Bob Real")
	seedMemberSearchMember(t, spaceId, verifiedUID, 0, 1)

	seedMemberSearchUser(t, placeholderUID, "", "s-carol", "", "")
	seedMemberSearchMember(t, spaceId, placeholderUID, 0, 1)

	list, err := testSpaceDB.searchMembers(spaceId, "", 1, 50)
	require.NoError(t, err)

	byUID := map[string]string{}
	for _, m := range list {
		byUID[m.UID] = m.Name
	}

	assert.Equal(t, "Alice", byUID[namedUID])
	assert.Equal(t, "Bob Real", byUID[verifiedUID])
	assert.Equal(t, "User-"+placeholderUID[:6], byUID[placeholderUID])
}
