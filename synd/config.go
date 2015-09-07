package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type configFromTOML struct {
	Master     string   `toml:"master"`
	Slaves     []string `toml:"slaves"`
	NetFilter  []string `toml:"netfilter"`
	TierFilter []string `toml:"tierfilter"`
	Port       int      `toml:"port"`
	RingDir    string   `toml:"ringdir"`
	CertFile   string   `toml:"certfile"`
	KeyFile    string   `toml:"keyfile"`
	UseTLS     string   `toml:"tls"`
	OPLog      string   `toml:"oplog"`
}

type Config struct {
	Master     bool
	SlavesArg  []string
	NetFilter  []string
	TierFilter []string
	Port       int
	RingDir    string
	CertFile   string
	KeyFile    string
	UseTLS     bool
	OPLog      string
	Slaves     []*RingSlave
}

var (
	DefaultNetFilter  = []string{"10.0.0.0/8", "192.168.0.0/16"} //need to pull from conf
	DefaultTierFilter = []string{"z.*"}
)

func parseSlaveAddrs(slaveAddrs []string) []*RingSlave {
	slaves := make([]*RingSlave, len(slaveAddrs))
	for i, v := range slaveAddrs {
		slaves[i] = &RingSlave{
			status: false,
			addr:   v,
		}
	}
	return slaves
}

func loadConfig(f string) (*Config, error) {
	c := &Config{}
	var ct configFromTOML
	if _, err := toml.DecodeFile(f, &ct); err != nil {
		log.Println(err)
	}

	// first we apply any values we got via the toml config
	if ct.Master == "true" {
		c.Master = true
	}
	if ct.Slaves != nil && len(ct.Slaves) >= 1 {
		c.SlavesArg = ct.Slaves
	}
	if ct.NetFilter != nil && len(ct.NetFilter) >= 1 {
		c.NetFilter = ct.NetFilter
	}
	if ct.TierFilter != nil && len(ct.TierFilter) >= 1 {
		c.TierFilter = ct.TierFilter
	}
	if ct.Port != 0 {
		c.Port = ct.Port
	}
	if ct.RingDir != "" {
		c.RingDir = ct.RingDir
	}
	if ct.CertFile != "" {
		c.CertFile = ct.CertFile
	}
	if ct.KeyFile != "" {
		c.KeyFile = ct.KeyFile
	}
	if ct.UseTLS == "true" {
		c.UseTLS = true
	}
	if ct.OPLog != "" {
		c.OPLog = ct.OPLog
	}

	//Now override all with passed cli flags
	flag.BoolVar(&c.Master, "master", c.Master, "Run as master")
	slavearg := flag.String("slaves", "", "comma sep list of ring slave addresses")
	if *slavearg != "" {
		c.SlavesArg = strings.Split(*slavearg, ",")
	}
	flag.StringVar(&c.RingDir, "ring_dir", c.RingDir, "ring dir to use for storage")
	flag.IntVar(&c.Port, "port", c.Port, "Port to bind")
	flag.StringVar(&c.CertFile, "cert_file", c.CertFile, "TLS Cert file")
	flag.StringVar(&c.KeyFile, "key_file", c.KeyFile, "TLS Key file")
	flag.BoolVar(&c.UseTLS, "tls", c.UseTLS, "use tls")
	flag.Parse()

	//Now check for/set defaults
	c.Slaves = parseSlaveAddrs(c.SlavesArg)
	if c.NetFilter == nil {
		c.NetFilter = DefaultNetFilter
		log.Println("Using default net filter:", c.NetFilter)
	}
	if c.TierFilter == nil {
		c.TierFilter = DefaultTierFilter
		log.Println("Using default tier filter:", c.TierFilter)
	}
	if c.RingDir == "" {
		c.RingDir = "/etc/ort/ring"
		d, err := os.Stat(c.RingDir)
		if err != nil {
			return nil, fmt.Errorf("Error check on ringdir: %s", err)
		}
		if !d.IsDir() {
			return nil, fmt.Errorf("Error check on ringdir. ringdir is not a dir")
		}
	}
	if c.Port == 0 {
		c.Port = 8443
	}
	if c.CertFile == "nil" {
		c.CertFile = "server.crt"
	}
	if c.KeyFile == "nil" {
		c.KeyFile = "server.key"
	}
	if ct.OPLog == "" {
		c.OPLog = filepath.Join(c.RingDir, "synd.log")
	}

	return c, nil
}
