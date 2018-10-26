// bring up two nodes and connect them
package main

import (
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	demo "../common"
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

	// get the node instance of the second server
	node2 := srv2.Self()

	// add it as a peer to the first node
	// the connection and crypto handshake will be performed automatically
	srv1.AddPeer(node2)

	// wait for the connection to complete
	time.Sleep(time.Millisecond * 100)

	// inspect the results
	demo.LogInfo("after add", "node1's peers", srv1.Peers(), "node2's peers", srv2.Peers())

	// stop the servers
	srv1.Stop()
	srv2.Stop()
}
