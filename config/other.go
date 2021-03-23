package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ghodss/yaml"
	"github.com/ipfs/go-log/v2"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
)

const (
	filesPerm = 0600
	dirsPerm  = 0700
)

var logger = log.Logger("awl/config")

var (
	executableDir string
)

func init() {
	ex, err := os.Executable()
	if err != nil {
		logger.DPanicf("find executable: %v", err)
		return
	}
	executableDir = filepath.Dir(ex)
}

func CalcAppDataDir() string {
	if executableDir != "" {
		configPath := filepath.Join(executableDir, AppConfigFilename)
		if _, err := os.Stat(configPath); err == nil {
			return executableDir
		}
	}
	return ""
}

func NewConfig() *Config {
	conf := &Config{}
	setDefaults(conf)
	return conf
}

func LoadConfig() (*Config, error) {
	dataDir := CalcAppDataDir()
	configPath := filepath.Join(dataDir, AppConfigFilename)
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	conf := new(Config)
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return nil, err
	}
	setDefaults(conf)
	return conf, nil
}

func setDefaults(conf *Config) {
	// P2pNode
	if len(conf.P2pNode.ListenAddresses) == 0 {
		multiaddrs := make([]multiaddr.Multiaddr, 0, 4)
		for _, s := range []string{
			"/ip4/0.0.0.0/tcp/0",
			"/ip6/::/tcp/0",
			"/ip4/0.0.0.0/udp/0",
			"/ip4/0.0.0.0/udp/0/quic",
			"/ip6/::/udp/0",
			"/ip6/::/udp/0/quic",
		} {
			addr, err := multiaddr.NewMultiaddr(s)
			if err != nil {
				logger.DPanicf("parse multiaddr: %v", err)
			}
			multiaddrs = append(multiaddrs, addr)
		}
		conf.SetListenAddresses(multiaddrs)
	}
	if conf.P2pNode.BootstrapPeers == nil {
		conf.P2pNode.BootstrapPeers = make([]string, 0)
	}
	if conf.P2pNode.DHTProtocolPrefix == "" {
		conf.P2pNode.DHTProtocolPrefix = dht.DefaultPrefix
	}

	// Other
	if conf.LoggerLevel == "" {
		conf.LoggerLevel = "debug"
	}
	if conf.HttpListenAddress == "" {
		conf.HttpListenAddress = "127.0.0.1:" + strconv.Itoa(DefaultHTTPPort)
	}
}
