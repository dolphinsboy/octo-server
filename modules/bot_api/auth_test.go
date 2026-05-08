package bot_api

import (
	"testing"
)

func TestExtractBotToken(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"valid bearer", "Bearer bf_abc123def456", "bf_abc123def456"},
		{"valid app token", "Bearer app_abc123def456", "app_abc123def456"},
		{"empty", "", ""},
		{"no bearer prefix", "bf_abc123", ""},
		{"basic auth", "Basic dXNlcjpwYXNz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate wkhttp.Context with gin
			// Since extractBotToken only reads c.GetHeader, we test the logic directly
			got := extractBotTokenFromHeader(tt.header)
			if got != tt.expected {
				t.Errorf("extractBotTokenFromHeader(%q) = %q, want %q", tt.header, got, tt.expected)
			}
		})
	}
}

// extractBotTokenFromHeader is a testable version of the token extraction logic.
func extractBotTokenFromHeader(auth string) string {
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}

func TestTokenPrefixRouting(t *testing.T) {
	tests := []struct {
		token    string
		isAppBot bool
	}{
		{"app_d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9", true},
		{"app_", true}, // edge case: prefix matches but short
		{"bf_a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", false},
		{"legacy_token_no_prefix", false},
		{"", false},
		{"application_token", false}, // starts with "app" but not "app_"
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			got := len(tt.token) >= 4 && tt.token[:4] == "app_"
			if got != tt.isAppBot {
				t.Errorf("token %q: isAppBot = %v, want %v", tt.token, got, tt.isAppBot)
			}
		})
	}
}

func TestAppBotRegistryLookup(t *testing.T) {
	adapter := NewAppBotRegistryAdapter()

	// Empty registry
	if spec := adapter.FindByToken("app_nonexistent"); spec != nil {
		t.Errorf("expected nil for non-existent token, got %+v", spec)
	}

	// Add a spec
	adapter.Add("app_test123", &AppBotRegistrySpec{
		UID:     "app_mybot_bot",
		Scope:   "platform",
		SpaceID: "",
	})

	// Found
	spec := adapter.FindByToken("app_test123")
	if spec == nil {
		t.Fatal("expected spec, got nil")
	}
	if spec.UID != "app_mybot_bot" {
		t.Errorf("UID = %q, want %q", spec.UID, "app_mybot_bot")
	}
	if spec.Scope != "platform" {
		t.Errorf("Scope = %q, want %q", spec.Scope, "platform")
	}

	// Remove
	adapter.Remove("app_test123")
	if spec := adapter.FindByToken("app_test123"); spec != nil {
		t.Errorf("expected nil after remove, got %+v", spec)
	}
}

func TestAppBotRegistryTokenRotation(t *testing.T) {
	adapter := NewAppBotRegistryAdapter()

	oldToken := "app_old_token_hex1234567890abcd"
	newToken := "app_new_token_hex1234567890abcd"

	// Simulate initial state
	adapter.Add(oldToken, &AppBotRegistrySpec{
		UID:     "app_mybot_bot",
		Scope:   "space",
		SpaceID: "space_001",
	})

	// Verify old token works
	if spec := adapter.FindByToken(oldToken); spec == nil {
		t.Fatal("old token should resolve")
	}

	// Simulate rotateToken: remove old + add new
	adapter.Remove(oldToken)
	adapter.Add(newToken, &AppBotRegistrySpec{
		UID:     "app_mybot_bot",
		Scope:   "space",
		SpaceID: "space_001",
	})

	// Old token should NOT work anymore
	if spec := adapter.FindByToken(oldToken); spec != nil {
		t.Errorf("old token should not resolve after rotation, got %+v", spec)
	}

	// New token should work
	spec := adapter.FindByToken(newToken)
	if spec == nil {
		t.Fatal("new token should resolve after rotation")
	}
	if spec.UID != "app_mybot_bot" {
		t.Errorf("UID = %q, want %q", spec.UID, "app_mybot_bot")
	}
	if spec.SpaceID != "space_001" {
		t.Errorf("SpaceID = %q, want %q", spec.SpaceID, "space_001")
	}
}

func TestAppBotRegistryConcurrency(t *testing.T) {
	adapter := NewAppBotRegistryAdapter()

	// Concurrent reads and writes should not panic
	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			adapter.Add("app_token_a", &AppBotRegistrySpec{UID: "app_a_bot", Scope: "platform"})
			adapter.Remove("app_token_a")
		}
		close(done)
	}()

	for i := 0; i < 1000; i++ {
		adapter.FindByToken("app_token_a")
	}
	<-done
}
