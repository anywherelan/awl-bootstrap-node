package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/ghodss/yaml"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mr-tron/base58/base58"
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
		peerID            peer.ID
	}
	P2pNode struct {
		PeerID   string
		Identity string

		BootstrapPeers  []string
		ListenAddresses []string

		ExchangeIdentityWithPeersInBackground bool
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
	identity := base58.Encode(by)

	c.P2pNode.Identity = identity
	c.P2pNode.PeerID = id.String()
	c.peerID = id
	c.save()
	c.Unlock()
}

func (c *Config) PrivKey() []byte {
	c.RLock()
	defer c.RUnlock()

	if c.P2pNode.Identity == "" {
		return nil
	}
	b, err := base58.Decode(c.P2pNode.Identity)
	if err != nil {
		return nil
	}
	return b
}

func (c *Config) GetBootstrapPeers() []peer.AddrInfo {
	c.RLock()
	allMultiaddrs := make([]multiaddr.Multiaddr, 0, len(c.P2pNode.BootstrapPeers))
	for _, val := range c.P2pNode.BootstrapPeers {
		newMultiaddr, err := multiaddr.NewMultiaddr(val)
		if err != nil {
			logger.Warnf("invalid bootstrap multiaddr from config: %v", err)
			continue
		}
		peerInfo, err := peer.AddrInfoFromP2pAddr(newMultiaddr)
		if err == nil && peerInfo.ID == c.peerID {
			continue
		}

		allMultiaddrs = append(allMultiaddrs, newMultiaddr)
	}
	c.RUnlock()

	addrInfos, err := peer.AddrInfosFromP2pAddrs(allMultiaddrs...)
	if err != nil {
		logger.Warn("invalid one or more bootstrap addr info from config")
	}

	return addrInfos
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
	err = os.WriteFile(path, data, filesPerm)
	if err != nil {
		logger.DPanicf("Save config: %v", err)
	}
}
