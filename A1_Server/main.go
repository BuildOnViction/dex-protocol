package main

import (
	"github.com/ethereum/go-ethereum/crypto"

	demo "github.com/tomochain/orderbook/common"
)

// the best way to learn is stick to the code, simple
// and change the code to match with tomochain core
// and maybe need to run in the same package or export method for testing

func main() {

	// make a new private key
	privkey, err := crypto.GenerateKey()
	if err != nil {
		demo.LogCrit("Generate private key failed", "err", err)
	}

	srv := demo.NewServer(privkey, "DEX Demo 1", "1.0.0", nil, 0)

	// attempt to start the server
	err = srv.Start()
	if err != nil {
		demo.LogCrit("Start p2p.Server failed", "err", err)
	}

	// inspect the resulting values
	nodeinfo := srv.NodeInfo()
	demo.LogInfo("server started", "enode", nodeinfo.Enode, "name", nodeinfo.Name, "ID", nodeinfo.ID, "IP", nodeinfo.IP)

	// bring down the server
	srv.Stop()
}
