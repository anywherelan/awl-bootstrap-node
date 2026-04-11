package config

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func TestSetDefaults(t *testing.T) {
	conf := NewConfig()

	if len(conf.P2pNode.ListenAddresses) != 4 {
		t.Fatalf("expected 4 default listen addresses, got %d", len(conf.P2pNode.ListenAddresses))
	}
	if len(conf.P2pNode.BootstrapPeers) != 0 {
		t.Fatal("BootstrapPeers should be empty")
	}
	if conf.LoggerLevel != "info" {
		t.Errorf("default LoggerLevel = %q, want info", conf.LoggerLevel)
	}
	if conf.HttpListenAddress == "" {
		t.Error("HttpListenAddress should have a default")
	}
}

func TestNewExampleConfig(t *testing.T) {
	conf := NewExampleConfig()

	if len(conf.P2pNode.BootstrapPeers) == 0 {
		t.Fatal("example config should include default bootstrap peers")
	}
	if len(conf.GetListenAddresses()) == 0 {
		t.Fatal("example config should have listen addresses")
	}
}

func TestSetIdentityAndPrivKey(t *testing.T) {
	t.Chdir(t.TempDir())

	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}

	conf := NewConfig()
	conf.SetIdentity(priv, pid)

	if conf.P2pNode.PeerID != pid.String() {
		t.Errorf("PeerID not stored: got %q, want %q", conf.P2pNode.PeerID, pid.String())
	}
	if conf.P2pNode.Identity == "" {
		t.Error("Identity should be set after SetIdentity")
	}

	raw := conf.PrivKey()
	if raw == nil {
		t.Fatal("PrivKey returned nil after SetIdentity")
	}
	wantRaw, err := priv.Raw()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, wantRaw) {
		t.Error("PrivKey round-trip produced different bytes")
	}
}

func TestSaveLoadConfigRoundTrip(t *testing.T) {
	t.Chdir(t.TempDir())

	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}

	orig := NewExampleConfig()
	orig.SetIdentity(priv, pid)

	if err := SaveConfig(orig, AppConfigFilename); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if loaded.P2pNode.PeerID != orig.P2pNode.PeerID {
		t.Errorf("PeerID mismatch: got %q, want %q", loaded.P2pNode.PeerID, orig.P2pNode.PeerID)
	}
	if loaded.P2pNode.Identity != orig.P2pNode.Identity {
		t.Error("Identity not preserved through save/load")
	}
	if len(loaded.P2pNode.BootstrapPeers) != len(orig.P2pNode.BootstrapPeers) {
		t.Errorf("BootstrapPeers count mismatch: got %d, want %d",
			len(loaded.P2pNode.BootstrapPeers), len(orig.P2pNode.BootstrapPeers))
	}
	if len(loaded.GetListenAddresses()) != len(orig.GetListenAddresses()) {
		t.Errorf("ListenAddresses count mismatch: got %d, want %d",
			len(loaded.GetListenAddresses()), len(orig.GetListenAddresses()))
	}

	// GetBootstrapPeers should parse without errors for the known-good default set.
	infos := loaded.GetBootstrapPeers()
	if len(infos) == 0 {
		t.Error("GetBootstrapPeers returned no peers for example config")
	}
}

func TestPeerstoreDirIsRelativeWhenNoAppDataDir(t *testing.T) {
	t.Chdir(t.TempDir())
	conf := NewConfig()
	// With no config.yaml beside the test binary, CalcAppDataDir returns "".
	want := filepath.Join("", DhtPeerstoreDataDirectory)
	if got := conf.PeerstoreDir(); got != want {
		t.Errorf("PeerstoreDir = %q, want %q", got, want)
	}
}
