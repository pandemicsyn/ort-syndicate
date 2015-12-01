package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"log"
	"net"

	"github.com/BurntSushi/toml"
	pb "github.com/pandemicsyn/syndicate/api/proto"
	"github.com/pandemicsyn/syndicate/syndicate"
)

var (
	printVersionInfo = flag.Bool("version", false, "print version/build info")
)

var syndVersion string
var ringVersion string
var goVersion string
var buildDate string

/*
func newRingDistServer() *ringslave {
	s := new(ringslave)
	return s
}
*/

type RingSyndicate struct {
	sync.RWMutex
	active bool
	name   string
	config *syndicate.Config
	server *syndicate.Server
}

type RingSyndicates struct {
	sync.RWMutex
	Syndics []*RingSyndicate
}

type ClusterConfigs struct {
	ValueSyndicate *syndicate.Config
	GroupSyndicate *syndicate.Config
}

func launchSyndicates(rs *RingSyndicates) {
	rs.Lock()
	defer rs.Unlock()
	for k, _ := range rs.Syndics {
		rs.Syndics[k].Lock()
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", rs.Syndics[k].config.Port))
		if err != nil {
			log.Fatalln(err)
			return
		}
		var opts []grpc.ServerOption
		creds, err := credentials.NewServerTLSFromFile(rs.Syndics[k].config.CertFile, rs.Syndics[k].config.KeyFile)
		if err != nil {
			log.Fatalln("Error load cert or key:", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
		s := grpc.NewServer(opts...)

		if rs.Syndics[k].config.Master {
			pb.RegisterSyndicateServer(s, rs.Syndics[k].server)
			log.Println("Master starting up on", rs.Syndics[k].config.Port)
			s.Serve(l)
		} else {
			//pb.RegisterRingDistServer(s, newRingDistServer())
			//log.Printf("Starting ring slave up on %d...\n", cfg.Port)
			//s.Serve(l)
			log.Fatalln("Syndicate slaves not implemented yet")
		}
		rs.Syndics[k].Unlock()
	}
}

func main() {
	var err error
	configFile := "/etc/oort/syndicate.toml"
	if os.Getenv("SYNDICATE_CONFIG") != "" {
		configFile = os.Getenv("SYNDICATE_CONFIG")
	}
	flag.Parse()
	if *printVersionInfo {
		fmt.Println("syndicate-client:", syndVersion)
		fmt.Println("ring version:", ringVersion)
		fmt.Println("build date:", buildDate)
		fmt.Println("go version:", goVersion)
		return
	}
	rs := &RingSyndicates{}

	var tc map[string]syndicate.Config
	if _, err := toml.DecodeFile(configFile, &tc); err != nil {
		log.Fatalln(err)
	}
	for k, v := range tc {
		log.Println("Found config for", k)
		log.Println("Config:", v)
		syndic := &RingSyndicate{
			active: false,
			name:   k,
			config: &v,
		}
		syndic.server, err = syndicate.NewServer(&v, k)
		if err != nil {
			log.Fatalln(err)
		}
		rs.Syndics = append(rs.Syndics, syndic)
	}
	launchSyndicates(rs)
}
