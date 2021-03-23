package config

import (
	"encoding/hex"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/ghodss/yaml"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap/zapcore"
)

const (
	AppConfigFilename         = "config.yaml"
	DhtPeerstoreDataDirectory = "peerstore"
	DefaultHTTPPort           = 9090
)

type (
	Config struct {
		sync.RWMutex
		P2pNode           P2pNode
		LoggerLevel       string
		HttpListenAddress string
	}
	P2pNode struct {
		PeerID   string
		Identity string

		BootstrapPeers    []string
		ListenAddresses   []string
		DHTProtocolPrefix protocol.ID
	}
)

func (c *Config) Save() {
	c.RLock()
	c.save()
	c.RUnlock()
}

func (c *Config) SetIdentity(key crypto.PrivKey, id peer.ID) {
	c.Lock()
	by, _ := key.Raw()
	identity := hex.EncodeToString(by)

	c.P2pNode.Identity = identity
	c.P2pNode.PeerID = id.Pretty()
	c.save()
	c.Unlock()
}

func (c *Config) PrivKey() []byte {
	c.RLock()
	defer c.RUnlock()

	if c.P2pNode.Identity == "" {
		return nil
	}
	b, err := hex.DecodeString(c.P2pNode.Identity)
	if err != nil {
		return nil
	}
	return b
}

func (c *Config) GetBootstrapPeers() []multiaddr.Multiaddr {
	c.RLock()
	allMultiaddrs := make([]multiaddr.Multiaddr, 0, len(c.P2pNode.BootstrapPeers))
	for _, val := range c.P2pNode.BootstrapPeers {
		newMultiaddr, _ := multiaddr.NewMultiaddr(val)
		allMultiaddrs = append(allMultiaddrs, newMultiaddr)
	}
	c.RUnlock()

	result := make([]multiaddr.Multiaddr, 0, len(allMultiaddrs))
	resultMap := make(map[string]struct{}, len(allMultiaddrs))
	for _, maddr := range allMultiaddrs {
		if _, exists := resultMap[maddr.String()]; !exists {
			resultMap[maddr.String()] = struct{}{}
			result = append(result, maddr)
		}
	}
	return result
}

func (c *Config) SetListenAddresses(multiaddrs []multiaddr.Multiaddr) {
	c.Lock()
	result := make([]string, 0, len(multiaddrs))
	for _, val := range multiaddrs {
		result = append(result, val.String())
	}
	c.P2pNode.ListenAddresses = result
	c.Unlock()
}

func (c *Config) GetListenAddresses() []multiaddr.Multiaddr {
	c.RLock()
	result := make([]multiaddr.Multiaddr, 0, len(c.P2pNode.ListenAddresses))
	for _, val := range c.P2pNode.ListenAddresses {
		newMultiaddr, _ := multiaddr.NewMultiaddr(val)
		result = append(result, newMultiaddr)
	}
	c.RUnlock()
	return result
}

func (c *Config) Path() string {
	return AppConfigFilename
}

func (c *Config) PeerstoreDir() string {
	dataDir := CalcAppDataDir()
	peerstoreDir := filepath.Join(dataDir, DhtPeerstoreDataDirectory)
	return peerstoreDir
}
func (c *Config) LogLevel() zapcore.Level {
	level := c.LoggerLevel
	if c.LoggerLevel == "dev" {
		level = "debug"
	}
	var lvl zapcore.Level
	_ = lvl.Set(level)
	return lvl
}

func (c *Config) DevMode() bool {
	return c.LoggerLevel == "dev"
}

func (c *Config) save() {
	data, err := yaml.Marshal(c)
	if err != nil {
		logger.DPanicf("Marshal config: %v", err)
	}
	path := c.Path()
	err = ioutil.WriteFile(path, data, filesPerm)
	if err != nil {
		logger.DPanicf("Save config: %v", err)
	}
}
