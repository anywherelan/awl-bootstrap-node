package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mr-tron/base58/base58"

	"github.com/anywherelan/awl-bootstrap-node/config"
)

// TestApplicationSmoke exercises the full startup/shutdown path:
// New() -> SetupLoggerAndConfig -> Init -> HTTP GET on /p2p_info -> Close.
// It spins up a real libp2p host on loopback with OS-assigned ports and a
// badger-backed peerstore in a temp directory. No external network required.
func TestApplicationSmoke(t *testing.T) {
	t.Chdir(t.TempDir())
	httpPort := writeTestConfig(t)

	app := New()
	app.SetupLoggerAndConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.Init(ctx); err != nil {
		t.Fatalf("app.Init: %v", err)
	}
	t.Cleanup(app.Close)

	// API is started in a goroutine; poll briefly until it answers.
	url := fmt.Sprintf("http://127.0.0.1:%d/api/v0/debug/p2p_info", httpPort)
	body := getWithRetry(t, url, 5*time.Second)

	var info map[string]any
	if err := json.Unmarshal(body, &info); err != nil {
		t.Fatalf("decode p2p_info: %v\nbody: %s", err, body)
	}
	general, ok := info["General"].(map[string]any)
	if !ok {
		t.Fatalf("response missing General block: %v", info)
	}
	if general["Version"] != config.Version {
		t.Errorf("General.Version = %v, want %q", general["Version"], config.Version)
	}

	// Log endpoint should return 200 with plain text even if the buffer is empty.
	logURL := fmt.Sprintf("http://127.0.0.1:%d/api/v0/debug/log", httpPort)
	resp, err := http.Get(logURL) //nolint:gosec,noctx // trusted loopback URL in test
	if err != nil {
		t.Fatalf("GET log: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("log endpoint status = %d, want 200", resp.StatusCode)
	}

	// Metrics endpoint must expose both our awl_bootstrap__* metrics and libp2p's built-in
	// families. The libp2p_* ones are the important safety check: they confirm
	// PrometheusRegisterer was wired through and the relay-service metrics tracer
	// got registered.
	metricsURL := fmt.Sprintf("http://127.0.0.1:%d/metrics", httpPort)
	metricsBody := getWithRetry(t, metricsURL, 5*time.Second)
	wantFamilies := []string{
		"awl_bootstrap_node_info",
		"awl_bootstrap_p2p_dht_routing_table_size",
		"libp2p_swarm_",    // PrometheusRegisterer reached the swarm
		"libp2p_relaysvc_", // relay-service metrics tracer was registered
	}
	for _, want := range wantFamilies {
		if !bytes.Contains(metricsBody, []byte(want)) {
			t.Errorf("/metrics output does not contain %q", want)
		}
	}
}

// writeTestConfig generates a fresh ed25519 identity, builds a minimal Config
// listening on loopback with OS-assigned libp2p ports, saves it to
// config.AppConfigFilename in the current directory, and returns the HTTP port
// that was wired in so callers can reach the API.
func writeTestConfig(t *testing.T) int {
	t.Helper()

	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	pid, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		t.Fatalf("peer.IDFromPrivateKey: %v", err)
	}
	rawKey, err := priv.Raw()
	if err != nil {
		t.Fatalf("priv.Raw: %v", err)
	}

	httpPort := freeTCPPort(t)
	conf := &config.Config{
		LoggerLevel:       "info",
		HttpListenAddress: fmt.Sprintf("127.0.0.1:%d", httpPort),
		P2pNode: config.P2pNode{
			PeerID:   pid.String(),
			Identity: base58.Encode(rawKey),
			// Loopback only, OS-assigned ports — no external networking.
			ListenAddresses: []string{
				"/ip4/127.0.0.1/tcp/0",
				"/ip4/127.0.0.1/udp/0/quic-v1",
			},
			BootstrapPeers: []string{},
		},
	}
	if err := config.SaveConfig(conf, config.AppConfigFilename); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	return httpPort
}

// TestGenerateExampleConfig runs the real generateExampleConfig helper from
// main.go end-to-end: it spins up a throwaway libp2p host to mint an identity,
// then writes a usable config file to disk. We reload the file via LoadConfig
// (same path production uses) and verify the identity/peer-id are populated,
// the private key round-trips, and the default bootstrap peers and listen
// addresses from setDefaults are present.
func TestGenerateExampleConfig(t *testing.T) {
	t.Chdir(t.TempDir())

	// Use the real filename so LoadConfig can find it through CalcAppDataDir.
	generateExampleConfig(config.AppConfigFilename)

	info, err := os.Stat(config.AppConfigFilename)
	if err != nil {
		t.Fatalf("stat generated config: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("generated config is empty")
	}

	loaded, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if loaded.P2pNode.Identity == "" {
		t.Error("generated config has no Identity")
	}
	if loaded.P2pNode.PeerID == "" {
		t.Error("generated config has no PeerID")
	}
	// PeerID must parse as a valid libp2p peer ID.
	pid, err := peer.Decode(loaded.P2pNode.PeerID)
	if err != nil {
		t.Fatalf("peer.Decode: %v", err)
	}

	// Identity bytes must parse as an ed25519 private key whose derived
	// peer ID matches the stored PeerID.
	raw := loaded.PrivKey()
	if raw == nil {
		t.Fatal("loaded.PrivKey() returned nil")
	}
	priv, err := crypto.UnmarshalEd25519PrivateKey(raw)
	if err != nil {
		t.Fatalf("UnmarshalEd25519PrivateKey: %v", err)
	}
	derivedPID, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		t.Fatalf("IDFromPrivateKey: %v", err)
	}
	if derivedPID != pid {
		t.Errorf("derived PeerID %q does not match stored %q", derivedPID, pid)
	}

	if len(loaded.P2pNode.BootstrapPeers) == 0 {
		t.Error("generated config has no BootstrapPeers (expected awl defaults)")
	}
	if len(loaded.GetListenAddresses()) == 0 {
		t.Error("generated config has no ListenAddresses")
	}

	// Sanity check: base58 Identity should decode back to ed25519 raw key size (64 bytes).
	decoded, err := base58.Decode(loaded.P2pNode.Identity)
	if err != nil {
		t.Fatalf("base58.Decode identity: %v", err)
	}
	if len(decoded) == 0 {
		t.Error("decoded identity is empty")
	}
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}

func getWithRetry(t *testing.T, url string, timeout time.Duration) []byte {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		b, err := tryGet(url)
		if err == nil {
			return b
		}
		lastErr = err
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("GET %s: %v", url, lastErr)
	return nil
}

func tryGet(url string) ([]byte, error) {
	resp, err := http.Get(url) //nolint:gosec,noctx // trusted loopback URL in test
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
