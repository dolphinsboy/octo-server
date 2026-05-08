package app_bot

import (
	"testing"
)

func TestGenerateAppBotToken(t *testing.T) {
	token, err := generateAppBotToken()
	if err != nil {
		t.Fatalf("generateAppBotToken() error: %v", err)
	}
	if len(token) != 36 { // "app_" + 32 hex chars
		t.Errorf("token length = %d, want 36, token=%q", len(token), token)
	}
	if token[:4] != AppBotTokenPrefix {
		t.Errorf("token prefix = %q, want %q", token[:4], AppBotTokenPrefix)
	}

	// Uniqueness check
	token2, _ := generateAppBotToken()
	if token == token2 {
		t.Error("two generated tokens should not be equal")
	}
}

func TestAppBotUIDFormat(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"octo-butler", "app_octo-butler_bot"},
		{"my_bot", "app_my_bot_bot"},
		{"a", "app_a_bot"},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			uid := AppBotUIDPrefix + tt.id + AppBotUIDSuffix
			if uid != tt.expected {
				t.Errorf("UID = %q, want %q", uid, tt.expected)
			}
		})
	}
}

func TestIDPattern(t *testing.T) {
	valid := []string{"octo-butler", "a", "my_bot_123", "a0", "abc-def_ghi"}
	invalid := []string{"", "A", "UPPERCASE", "-start", "_start", "has space", "a!b", repeat30Plus()}

	for _, id := range valid {
		if !idPattern.MatchString(id) {
			t.Errorf("expected %q to be valid", id)
		}
	}
	for _, id := range invalid {
		if idPattern.MatchString(id) {
			t.Errorf("expected %q to be invalid", id)
		}
	}
}

// repeat30Plus returns a string longer than 30 chars
func repeat30Plus() string {
	return "abcdefghijklmnopqrstuvwxyz12345" // 31 chars
}

func TestReservedIDs(t *testing.T) {
	reserved := []string{"system", "filehelper", "botfather", "notification"}
	for _, id := range reserved {
		if !reservedIDs[id] {
			t.Errorf("%q should be reserved", id)
		}
	}

	notReserved := []string{"mybot", "octo-butler", "custom"}
	for _, id := range notReserved {
		if reservedIDs[id] {
			t.Errorf("%q should not be reserved", id)
		}
	}
}

func TestRegistryAddRemove(t *testing.T) {
	r := NewRegistry()

	spec := &AppBotSpec{
		ID:          "test-bot",
		UID:         "app_test-bot_bot",
		DisplayName: "Test Bot",
		Scope:       "platform",
		Token:       "app_token123",
	}

	// Add
	r.Add(spec)
	if got := r.FindByUID("app_test-bot_bot"); got == nil {
		t.Error("expected to find by UID")
	}
	if got := r.FindByID("test-bot"); got == nil {
		t.Error("expected to find by ID")
	}

	// Remove
	r.Remove("test-bot", "app_test-bot_bot")
	if got := r.FindByUID("app_test-bot_bot"); got != nil {
		t.Error("expected nil after remove by UID")
	}
	if got := r.FindByID("test-bot"); got != nil {
		t.Error("expected nil after remove by ID")
	}
}

func TestRegistryAtomicTokenRotation(t *testing.T) {
	r := NewRegistry()

	// Initial state
	r.Add(&AppBotSpec{
		ID:    "mybot",
		UID:   "app_mybot_bot",
		Token: "app_old_token",
	})

	// Simulate rotateToken: Remove + Add with new token
	r.Remove("mybot", "app_mybot_bot")
	r.Add(&AppBotSpec{
		ID:    "mybot",
		UID:   "app_mybot_bot",
		Token: "app_new_token",
	})

	// Verify
	got := r.FindByID("mybot")
	if got == nil {
		t.Fatal("expected to find bot after rotation")
	}
	if got.Token != "app_new_token" {
		t.Errorf("Token = %q, want %q", got.Token, "app_new_token")
	}
}

func TestStatusConstants(t *testing.T) {
	if StatusDraft != 0 {
		t.Errorf("StatusDraft = %d, want 0", StatusDraft)
	}
	if StatusPublished != 1 {
		t.Errorf("StatusPublished = %d, want 1", StatusPublished)
	}
	if StatusUnpublished != 2 {
		t.Errorf("StatusUnpublished = %d, want 2", StatusUnpublished)
	}
}
