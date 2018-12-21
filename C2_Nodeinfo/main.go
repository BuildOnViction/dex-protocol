// Different ways of accessing RPC API on a servicenode
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	demo "github.com/tomochain/orderbook/common"
)

var (
	p2pPort       = 30100
	ipcpath       = ".demo.ipc"
	datadirPrefix = ".data_"
)

func main() {
	// set up the service node
	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", p2pPort)
	cfg.IPCPath = ipcpath
	cfg.DataDir = fmt.Sprintf("%s%d", datadirPrefix, p2pPort)

	// create the node instance with the config
	stack, err := node.New(cfg)
	if err != nil {
		demo.LogCrit("ServiceNode create fail", "err", err)
	}

	// start the node
	err = stack.Start()
	if err != nil {
		demo.LogCrit("ServiceNode start fail", "err", err)
	}
	defer os.RemoveAll(stack.DataDir())

	// get the info directly via the p2p server object
	p2pserver := stack.Server()
	localnodeinfo := p2pserver.NodeInfo()
	demo.LogInfo("Nodeinfo from p2p.Server", "enode", localnodeinfo.Enode, "IP", localnodeinfo.IP, "ID", localnodeinfo.ID, "listening address", localnodeinfo.ListenAddr)

	// get the nodeinfo via ServiceNode IPC
	localnodeinfo = &p2p.NodeInfo{}
	// attach client from node
	rpcclient, err := stack.Attach()
	// call method
	err = rpcclient.Call(&localnodeinfo, "admin_nodeInfo")
	if err != nil {
		demo.LogCrit("Could not get rpcclient via p2p.Server", "err", err)

	}
	demo.LogInfo("Nodeinfo from IPC via ServiceNode", "enode", localnodeinfo.Enode, "IP", localnodeinfo.IP, "ID", localnodeinfo.ID, "listening address", localnodeinfo.ListenAddr)

	// get the nodeinfo via external IPC
	rpcclient, err = rpc.Dial(filepath.Join(cfg.DataDir, cfg.IPCPath))
	if err != nil {
		demo.LogCrit("Could not get rpcclient via p2p.Server", "err", err)
	}
	localnodeinfo = &p2p.NodeInfo{}
	rpcclient, err = stack.Attach()
	err = rpcclient.Call(&localnodeinfo, "admin_nodeInfo")
	demo.LogInfo("Nodeinfo from IPC via external call", "enode", localnodeinfo.Enode, "IP", localnodeinfo.IP, "ID", localnodeinfo.ID, "listening address", localnodeinfo.ListenAddr)

	err = stack.Stop()
	if err != nil {
		demo.LogCrit("Node stop fail", "err", err)
	}
}
