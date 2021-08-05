//
// Copyright © 2021 Bedag Informatik AG
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package main

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	api "github.com/osrg/gobgp/api"
	gobgp "github.com/osrg/gobgp/pkg/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config config for goblackhole
type Config struct {
	Peers     []Peer
	Blocklist string
	Local_as  uint32
	Local_id  string
	Listen    int32
	NextHop   string
	LogLevel  string
	Interval  time.Duration
	Community []uint32
}

// Peer peer config
type Peer struct {
	Remote_as uint32
	Local_as  uint32
	Remote_ip string
}

var (
	cfg   Config           // config
	s     *gobgp.BgpServer //gobgp server
	attrs []*any.Any       //default bgp attributes
)

func main() {
	cfg = Config{}

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/goblackhole/")

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	viper.SetDefault("local_as", "65003")
	viper.SetDefault("local_id", "127.0.0.1")
	viper.SetDefault("Listen", -1)
	viper.SetDefault("Blocklist", "https://raw.githubusercontent.com/stamparm/ipsum/master/ipsum.txt")
	viper.SetDefault("NextHop", "192.168.0.1")
	viper.SetDefault("Log", "Info")
	viper.SetDefault("Interval", 5*time.Second)
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	level, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = log.DebugLevel
	}
	log.SetLevel(level)

	s = gobgp.NewBgpServer()
	go s.Serve()

	// global configuration
	if err := s.StartBgp(context.Background(), &api.StartBgpRequest{
		Global: &api.Global{
			As:         cfg.Local_as,
			RouterId:   cfg.Local_id,
			ListenPort: cfg.Listen,
		},
	}); err != nil {
		log.Fatal(err)
	}

	// Adding OriginAttribute
	a1, _ := ptypes.MarshalAny(&api.OriginAttribute{
		Origin: 0,
	})

	// Adding NextopAttribute
	a2, _ := ptypes.MarshalAny(&api.NextHopAttribute{
		NextHop: cfg.NextHop,
	})
	// Adding Community
	var communities []uint32
	for _, community := range cfg.Community {
		communities = append(communities, cfg.Local_as<<16^community)
	}
	a3, _ := ptypes.MarshalAny(&api.CommunitiesAttribute{
		Communities: communities,
	})

	// ToDo: Add comunityAttribute
	attrs = []*any.Any{a1, a2, a3}

	// monitor the change of the peer state
	if err := s.MonitorPeer(context.Background(), &api.MonitorPeerRequest{}, func(p *api.Peer) { log.Info(p) }); err != nil {
		log.Fatal(err)
	}

	// neighbor configuration
	for _, peer := range cfg.Peers {
		n := &api.Peer{
			Conf: &api.PeerConf{
				NeighborAddress: peer.Remote_ip,
				PeerAs:          peer.Remote_as,
				LocalAs:         cfg.Local_as,
			},
		}
		err := s.AddPeer(context.Background(), &api.AddPeerRequest{Peer: n})
		if err != nil {
			log.Fatal(err)
		}
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer func() {
		cancel()
	}()

	if err := loopFile(ctx); err != nil {
		log.Fatal(err)
	}

}

// getIP Input is string return ip error if not a ip
func getIP(ip string) (net.IPNet, error) {

	ip = strings.ReplaceAll(strings.Split(ip, "	")[0], " ", "")
	if net.ParseIP(ip) == nil {
		// try if with split
		if net.ParseIP(strings.Split(ip, "	")[0]) != nil {

		} else if _, _, err := net.ParseCIDR(ip); err != nil {
			log.Debugf("Cannot Parse IP '%v'", ip)
		}
	}
	if !strings.Contains(ip, "/") {
		ip = ip + "/32"
	}
	_, cidr, err := net.ParseCIDR(ip)
	if err != nil {
		return net.IPNet{}, err
	}
	return *cidr, nil
}

// readFile read remote file and return ip list
func readFile() ([]net.IPNet, error) {
	var ips []net.IPNet
	buf := new(bytes.Buffer)
	resp, err := http.Get(cfg.Blocklist)
	if err != nil {
		log.Fatal(err)
	}

	// Read Body to buffer fail function if error
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return []net.IPNet{}, err
	}

	// Loop over all lines and add ips
	for _, m := range strings.Split(buf.String(), "\n") {
		ip, err := getIP(m)
		if err != nil {
			log.Debug(err)
			continue
		}
		// Add ip to slice
		ips = append(ips, ip)
	}

	return ips, nil

}

// addIPtoPeer adding IPs to Peers
func addIPtoPeer(a []net.IPNet) {
	var err error
	for _, ip := range a {
		log.Debugf("adding %v to peers", ip.String())

		_, mask := ip.Mask.Size()
		nlri, _ := ptypes.MarshalAny(&api.IPAddressPrefix{
			Prefix:    string(ip.IP.String()),
			PrefixLen: uint32(mask),
		})
		_, err = s.AddPath(context.Background(), &api.AddPathRequest{
			Path: &api.Path{
				Family: &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
				Nlri:   nlri,
				Pattrs: attrs,
			},
		})

		if err != nil {
			log.Fatal(err)
		}

	}
}

// delIPtoPeer removes IPs to Peers
func delIPtoPeer(d []net.IPNet) {
	var err error

	for _, ip := range d {
		log.Debugf("delete %v to peers", ip.String())

		_, mask := ip.Mask.Size()
		nlri, _ := ptypes.MarshalAny(&api.IPAddressPrefix{
			Prefix:    string(ip.IP.String()),
			PrefixLen: uint32(mask),
		})
		err = s.DeletePath(context.Background(), &api.DeletePathRequest{
			Path: &api.Path{
				Family: &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
				Nlri:   nlri,
				Pattrs: attrs,
			},
		})

		if err != nil {
			log.Fatal(err)
		}

	}
}

// getIPDiff return slice of IP which should be
// removed and slice of IP which should be added.
// First Input new slice second input old slice
func getIPDiff(n, o []net.IPNet) (diffadd, diffdel []net.IPNet) {

	m := make(map[string]bool)

	for _, item := range o {
		m[item.String()] = true
	}

	for _, item := range n {
		if _, ok := m[item.String()]; !ok {
			diffadd = append(diffadd, item)
		}
	}

	m = make(map[string]bool)

	for _, item := range n {
		m[item.String()] = true
	}

	for _, item := range o {
		if _, ok := m[item.String()]; !ok {
			diffdel = append(diffdel, item)
		}
	}
	return diffadd, diffdel
}

// loopFile Function to refresh file
func loopFile(ctx context.Context) error {
	var (
		oIPs []net.IPNet
		nIPs []net.IPNet
		aIPs []net.IPNet
		dIPs []net.IPNet
		err  error
	)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.Tick(cfg.Interval):
			log.Debug("time ticked ;) update blacklist")
			// Get new IPs
			nIPs, err = readFile()
			if err != nil {
				return err
			}
			// Get diff new/old ips
			aIPs, dIPs = getIPDiff(nIPs, oIPs)

			// Add IPs to Peer
			addIPtoPeer(aIPs)

			// Remove Ips to Peer
			delIPtoPeer(dIPs)

			// Move New to Old and reset New
			oIPs = nIPs
			nIPs = []net.IPNet{}
		}
	}
}
