package space

import "testing"

func TestBuildChannelID(t *testing.T) {
	tests := []struct {
		spaceID, peerID, want string
	}{
		{"", "user123", "user123"},
		{"sp1", "user123", "ssp1_user123"},
		{"42", "bot_abc", "s42_bot_abc"},
	}
	for _, tt := range tests {
		got := BuildChannelID(tt.spaceID, tt.peerID)
		if got != tt.want {
			t.Errorf("BuildChannelID(%q, %q) = %q, want %q", tt.spaceID, tt.peerID, got, tt.want)
		}
	}
}

func TestParseChannelID(t *testing.T) {
	tests := []struct {
		channelID, wantSpace, wantPeer string
	}{
		{"user123", "", "user123"},
		{"ssp1_user123", "sp1", "user123"},
		{"s42_bot_abc", "42", "bot_abc"},
		{"notspace", "", "notspace"},
	}
	for _, tt := range tests {
		gotSpace, gotPeer := ParseChannelID(tt.channelID)
		if gotSpace != tt.wantSpace || gotPeer != tt.wantPeer {
			t.Errorf("ParseChannelID(%q) = (%q, %q), want (%q, %q)",
				tt.channelID, gotSpace, gotPeer, tt.wantSpace, tt.wantPeer)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	spaceID, peerID := "myspace", "user456"
	channelID := BuildChannelID(spaceID, peerID)
	gotSpace, gotPeer := ParseChannelID(channelID)
	if gotSpace != spaceID || gotPeer != peerID {
		t.Errorf("roundtrip failed: got (%q, %q), want (%q, %q)", gotSpace, gotPeer, spaceID, peerID)
	}
}
