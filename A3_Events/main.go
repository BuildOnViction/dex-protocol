// get notified when the peer connection has been completed
package main

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"

	demo "../common"
)

var (
	quitC = make(chan bool)
)

func main() {

	// we need private keys for both servers
	privkey1, err := crypto.GenerateKey()
	if err != nil {
		demo.LogCrit("Generate private key #1 failed", "err", err)
	}
	privkey2, err := crypto.GenerateKey()
	if err != nil {
		demo.LogCrit("Generate private key #2 failed", "err", err)
	}

	// set up the two servers
	srv1 := demo.NewServer(privkey1, "foo", "42", nil, 0)
	err = srv1.Start()
	if err != nil {
		demo.LogCrit("Start p2p.Server #1 failed", "err", err)
	}

	srv2 := demo.NewServer(privkey2, "bar", "666", nil, 31234)
	err = srv2.Start()
	if err != nil {
		demo.LogCrit("Start p2p.Server #2 failed", "err", err)
	}

	// set up the event subscription on the first server
	eventC := make(chan *p2p.PeerEvent)
	sub1 := srv1.SubscribeEvents(eventC)

	// listen for events in a go routine
	go func() {
		peerevent := <-eventC
		demo.LogInfo("Received peer Event", "type", peerevent.Type, "peer", peerevent.Peer)
		quitC <- true
	}()

	// get the node instance of the second server
	node2 := srv2.Self()

	// add it as a peer to the first node
	// the connection and crypto handshake will be performed automatically
	srv1.AddPeer(node2)

	// receives when the event is received, by waiting for quitC event
	<-quitC

	// inspect the results
	demo.LogInfo("After Add", "node1's peers", srv1.Peers(), "node2's peers", srv2.Peers())

	// terminate subscription
	sub1.Unsubscribe()

	// stop the servers
	srv1.Stop()
	srv2.Stop()
}
