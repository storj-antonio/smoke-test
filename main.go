package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-ping/ping"
	"go.uber.org/zap"
	"storj.io/common/identity"
	"storj.io/common/pb"
	"storj.io/common/peertls/tlsopts"
	"storj.io/common/rpc"
	"storj.io/common/storj"
	"storj.io/lib/uplink"
	"storj.io/private/memory"
)

var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

func NodeIDFromString(s string) storj.NodeID {
	Id, _ := storj.NodeIDFromString(s)
	return Id
}

func main() {
	ctx := context.Background()

	ident, err := identity.NewFullIdentity(ctx, identity.NewCAOptions{
		Difficulty:  9,
		Concurrency: 1,
	})
	if err != nil {
		return
	}

	cfg := &uplink.Config{}
	cfg.Volatile.MaxInlineSize = 4 * memory.KiB
	cfg.Volatile.MaxMemory = 4 * memory.MiB
	cfg.Volatile.Log = zap.NewNop()
	cfg.Volatile.DialTimeout = 20 * time.Second

	tlsConfig := tlsopts.Config{
		UsePeerCAWhitelist:  !cfg.Volatile.TLS.SkipPeerCAWhitelist,
		PeerCAWhitelistPath: cfg.Volatile.TLS.PeerCAWhitelistPath,
		PeerIDVersions:      "0",
	}

	tlsOptions, err := tlsopts.NewOptions(ident, tlsConfig, nil)
	if err != nil {
		return
	}

	dialer := rpc.NewDefaultDialer(tlsOptions)
	dialer.DialTimeout = cfg.Volatile.DialTimeout

	targets := []*pb.Node{
		{
			Id: NodeIDFromString("12PvuuRCUHBiqfDnmunUXfBhjGwGxgYVnEnibYXDdN9T1Pz3mqn"),
			Address: &pb.NodeAddress{
				Transport: pb.NodeTransport_TCP_TLS_GRPC,
				Address:   "symbiont.spdns.de:28970",
			},
		},
		{
			Id: NodeIDFromString("1WLfM29uVTNfMGfRdjbesJrAow1UvnqYgW94CLrxtcCXPQ7m4f"),
			Address: &pb.NodeAddress{
				Transport: pb.NodeTransport_TCP_TLS_GRPC,
				Address:   "storjnode000.dynv6.net:28967",
			},
		},
	}

	targets = targets[len(targets)-1:]

	for _, target := range targets {
		logger.Printf("%s : %s\n", target.Id.String(), target.Address)
		logger.Println("DialTime : PingTime")

		for w := 1; w <= 100; w++ {
			time.Sleep(3 * time.Second)
			start := time.Now()
			dialer.DialNode(ctx, target)
			dialtime := fmt.Sprint(time.Since(start))

			pinger, err := ping.NewPinger(strings.Split(target.Address.Address, ":")[0])
			if err != nil {
				panic(err)
			}

			pinger.Count = 1
			pinger.Run()                 // blocks until finished
			stats := pinger.Statistics() // get send/receive/rtt stats
			logger.Printf("%s : %s \n", dialtime, stats.AvgRtt)
		}
	}
}
