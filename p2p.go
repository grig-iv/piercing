package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
)

const (
	protocolID = "/piercing/0.0.0"
)

func p2pSetup(conf Config) {
	ctx := context.Background()

	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
	)
	if err != nil {
		log.Fatalf("Failed to create host: %v", err)
	}

	log.Printf("My Peer ID: %s", h.ID())
	log.Printf("My Addresses: %v", h.Addrs())

	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Bootstrapping DHT...")
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Go(func() {
			h.Connect(ctx, *peerinfo)
		})
	}
	wg.Wait()

	h.SetStreamHandler(protocolID, func(s network.Stream) {
		log.Printf("Received new stream from: %s", s.Conn().RemotePeer())

		kcv := make([]byte, 3)       // Length must be 3
		_, err = io.ReadFull(s, kcv) // Use ReadFull to ensure you get all 3 bytes

		for ring, key := range conf.Rings {
			keyHash := getHash(key)
			if slices.Equal(keyHash[:3], kcv) {
				fmt.Printf("Ring '%s' matched\n", ring)

				targets := make([]TargetMessage, 0)
				for _, target := range conf.Targets {
					if !slices.Contains(target.Rings, ring) {
						continue
					}

					state := Present

					fInfo, err := os.Stat(target.Path)
					if err != nil {
						if errors.Is(err, os.ErrNotExist) {
							state = Absent
						} else {
							log.Println("Stats", err)
							continue
						}
					}

					targets = append(targets, TargetMessage{
						TargetId:   target.Id,
						State:      state,
						LastChange: fInfo.ModTime(),
					})
				}

				err := json.NewEncoder(s).Encode(targets)
				if err != nil {
					log.Println("Error sending JSON:", err)
				}

				return
			}
		}

		s.Close()
	})

	routingDiscovery := routing.NewRoutingDiscovery(kademliaDHT)

	for ring, key := range conf.Rings {
		kcv := getHash(key)[:3]
		rendezvous := "piercing-" + hex.EncodeToString(getHash("salt"+key))
		log.Println(rendezvous)

		fmt.Printf("Advertise '%s' ring\n", ring)
		util.Advertise(ctx, routingDiscovery, rendezvous)

		go func() {
			for {
				peerChan, err := routingDiscovery.FindPeers(ctx, rendezvous)
				if err != nil {
					log.Fatal(err)
				}

				for peer := range peerChan {
					if peer.ID == h.ID() {
						continue
					}
					if h.Network().Connectedness(peer.ID) != network.Connected {
						s, err := h.NewStream(ctx, peer.ID, protocolID)
						if err != nil {
							log.Println("some peer err")
							continue
						}

						fmt.Printf("Connected to global peer: %s\n", peer.ID)
						fmt.Printf("Send ring KVC")

						if _, err := s.Write(kcv); err != nil {
							s.Reset()
							continue
						}

						var incoming []TargetMessage
						decoder := json.NewDecoder(s)
						if err := decoder.Decode(&incoming); err != nil {
							log.Println("Decode error:", err)
						}
						s.Close()
					}
				}
				time.Sleep(time.Second * 10)
			}
		}()
	}

	fmt.Println("Searching for peers globally via DHT...")

}
