// querying the p2p Server through RPC
package main

import (
	"net"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	demo "github.com/tomochain/orderbook/common"
)

func main() {

	// make a new private key
	privkey, err := crypto.GenerateKey()
	if err != nil {
		demo.LogCrit("Generate private key failed", "err", err)
	}

	srv := demo.NewServer(privkey, "foo", "42", nil, 0)

	// attempt to start the p2p server
	err = srv.Start()
	if err != nil {
		demo.LogCrit("Start p2p.Server failed", "err", err)
	}

	// set up the RPC server, so we can access all methods of srv server
	rpcsrv := rpc.NewServer()
	err = rpcsrv.RegisterName("foo", &srv)
	if err != nil {
		demo.LogCrit("Register API method(s) fail", "err", err)
	}

	// create IPC endpoint
	ipcpath := ".demo.ipc"
	ipclistener, err := net.Listen("unix", ipcpath)
	if err != nil {
		demo.LogCrit("IPC endpoint create fail", "err", err)
	}
	defer os.Remove(ipcpath)

	// mount RPC server on IPC endpoint
	go func() {
		err = rpcsrv.ServeListener(ipclistener)
		if err != nil {
			demo.LogCrit("Mount RPC on IPC fail", "err", err)
		}
	}()

	// create a IPC client
	rpcclient, err := rpc.Dial(ipcpath)
	if err != nil {
		demo.LogCrit("IPC dial fail", "err", err)
	}

	// call the RPC method
	var nodeinfo p2p.NodeInfo
	err = rpcclient.Call(&nodeinfo, "foo_nodeInfo")
	if err != nil {
		demo.LogCrit("RPC call fail", "err", err)
	}
	demo.LogInfo("server started", "enode", nodeinfo.Enode, "name", nodeinfo.Name, "ID", nodeinfo.ID, "IP", nodeinfo.IP)

	// bring down the servers
	rpcsrv.Stop()
	srv.Stop()
}
