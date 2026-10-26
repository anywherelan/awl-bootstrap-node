package metrics

import (
	"regexp"
	"strings"
	"sync"

	awlconfig "github.com/anywherelan/awl/config"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	"github.com/anywherelan/awl-bootstrap-node/config"
)

const (
	labelUnknown = "unknown"
	labelOther   = "other"
	labelDev     = "dev"

	appAwl          = "awl"
	appAwlBootstrap = "awl-bootstrap"

	labelNameApp     = "app"
	labelNameVersion = "version"
	labelNameOS      = "os"
	labelNameArch    = "arch"

	// A peer chooses its user agent, and every new label value is a new
	// series: these cap the distinct values of the labels taken from it.
	maxVersionLabels = 64
	maxOSLabels      = 16
	maxArchLabels    = 16
)

var (
	releaseVersionRe = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

	versionLabels = &labelLimiter{max: maxVersionLabels}
	osLabels      = &labelLimiter{max: maxOSLabels}
	archLabels    = &labelLimiter{max: maxArchLabels}
)

// clientLabels describe the client of a peer by the user agent it sent in
// identify.
type clientLabels struct {
	app     string
	version string
	os      string
	arch    string
}

// parseUserAgent understands the user agents of awl and of bootstrap nodes.
// Any other non-empty user agent is app "other".
func parseUserAgent(userAgent string) clientLabels {
	labels := clientLabels{app: labelUnknown, version: labelUnknown, os: labelUnknown, arch: labelUnknown}
	if userAgent == "" {
		return labels
	}

	switch {
	case strings.HasPrefix(userAgent, awlconfig.UserAgentPrefix):
		goos, goarch := awlconfig.SystemInfoFromUserAgent(userAgent)
		labels.app = appAwl
		labels.version = versionLabel(awlconfig.VersionFromUserAgent(userAgent))
		labels.os = osLabels.value(goos)
		labels.arch = archLabels.value(goarch)
	case strings.HasPrefix(userAgent, config.UserAgentPrefix):
		labels.app = appAwlBootstrap
		labels.version = versionLabel(awlconfig.VersionFromUserAgent(userAgent))
	default:
		labels.app = labelOther
	}

	return labels
}

// versionLabel keeps release versions and calls everything else, such as
// "dev" or "v0.19.0-57-gdb68033b" from git describe, a dev build.
func versionLabel(version string) string {
	if !releaseVersionRe.MatchString(version) {
		return labelDev
	}
	return versionLabels.value(version)
}

func peerUserAgent(ps peerstore.Peerstore, id peer.ID) string {
	agent, err := ps.Get(id, "AgentVersion")
	if err != nil {
		return ""
	}
	userAgent, _ := agent.(string)
	return userAgent
}

// labelLimiter lets through at most max distinct values and replaces the
// rest with "other". An empty value is "unknown".
type labelLimiter struct {
	mu   sync.Mutex
	seen map[string]struct{}
	max  int
}

func (l *labelLimiter) value(v string) string {
	if v == "" {
		return labelUnknown
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.seen[v]; ok {
		return v
	}
	if len(l.seen) >= l.max {
		return labelOther
	}
	if l.seen == nil {
		l.seen = make(map[string]struct{})
	}
	l.seen[v] = struct{}{}
	return v
}
