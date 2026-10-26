package metrics

import (
	"testing"

	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
)

func TestParseUserAgent(t *testing.T) {
	tests := []struct {
		userAgent string
		want      clientLabels
	}{
		{"", clientLabels{"unknown", "unknown", "unknown", "unknown"}},
		{"awl/linux-amd64/v0.20.0", clientLabels{"awl", "v0.20.0", "linux", "amd64"}},
		{"awl/android-arm64/v0.19.0", clientLabels{"awl", "v0.19.0", "android", "arm64"}},
		{"awl/linux-amd64/dev", clientLabels{"awl", "dev", "linux", "amd64"}},
		{"awl/windows-amd64/v0.19.0-57-gdb68033b", clientLabels{"awl", "dev", "windows", "amd64"}},
		{"awl/linux/v0.20.0", clientLabels{"awl", "v0.20.0", "linux", "unknown"}},
		{"awl-bootstrap/v0.10.0", clientLabels{"awl-bootstrap", "v0.10.0", "unknown", "unknown"}},
		{"go-libp2p/0.50.0", clientLabels{"other", "unknown", "unknown", "unknown"}},
	}
	for _, tt := range tests {
		if got := parseUserAgent(tt.userAgent); got != tt.want {
			t.Errorf("parseUserAgent(%q) = %+v, want %+v", tt.userAgent, got, tt.want)
		}
	}
}

func TestLabelLimiter(t *testing.T) {
	limiter := &labelLimiter{max: 2}
	for _, tt := range []struct{ value, want string }{
		{"v0.1.0", "v0.1.0"},
		{"v0.2.0", "v0.2.0"},
		{"v0.3.0", "other"},
		{"v0.1.0", "v0.1.0"},
		{"", "unknown"},
	} {
		if got := limiter.value(tt.value); got != tt.want {
			t.Errorf("value(%q) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestMessageTypeLabel(t *testing.T) {
	if got := messageTypeLabel(pb.Message_FIND_NODE); got != "FIND_NODE" {
		t.Errorf("messageTypeLabel(FIND_NODE) = %q", got)
	}
	if got := messageTypeLabel(pb.Message_MessageType(99)); got != "other" {
		t.Errorf("messageTypeLabel(99) = %q, want other", got)
	}
}
