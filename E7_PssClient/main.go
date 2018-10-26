// pss RPC routed over swarm
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/swarm/pss"
	pssclient "github.com/ethereum/go-ethereum/swarm/pss/client"

	demo "../common"
)

// simple ping and receive protocol
var (
	topic = pss.PingTopic
	quitC = make(chan struct{})
)

func main() {

	// create two nodes
	leftStack, err := demo.NewServiceNode(demo.P2pPort, 0, demo.WSDefaultPort, "pss")
	if err != nil {
		demo.LogCrit(err.Error())
	}
	rightStack, err := demo.NewServiceNode(demo.P2pPort+1, 0, demo.WSDefaultPort+1, "pss")
	if err != nil {
		demo.LogCrit(err.Error())
	}

	// register the pss activated bzz services
	leftSvc := demo.NewSwarmService(leftStack, demo.BzzDefaultPort)
	err = leftStack.Register(leftSvc)
	if err != nil {
		demo.LogCrit("servicenode 'left' pss register fail", "err", err)
	}

	rightSvc := demo.NewSwarmService(rightStack, demo.BzzDefaultPort+1)
	err = rightStack.Register(rightSvc)
	if err != nil {
		demo.LogCrit("servicenode 'right' pss register fail", "err", err)
	}

	// start the nodes
	err = leftStack.Start()
	if err != nil {
		demo.LogCrit("servicenode start failed", "err", err)
	}
	defer os.RemoveAll(leftStack.DataDir())
	err = rightStack.Start()
	if err != nil {
		demo.LogCrit("servicenode start failed", "err", err)
	}
	defer os.RemoveAll(rightStack.DataDir())

	// connect the nodes to the middle
	leftStack.Server().AddPeer(rightStack.Server().Self())

	// get the rpc clients
	leftRPCClient, err := leftStack.Attach()
	rightRPCClient, err := rightStack.Attach()

	// get the public keys
	var leftPubKey string
	err = leftRPCClient.Call(&leftPubKey, "pss_getPublicKey")
	if err != nil {
		demo.LogCrit("pss get pubkey fail", "err", err)
	}
	var rightPubKey string
	err = rightRPCClient.Call(&rightPubKey, "pss_getPublicKey")
	if err != nil {
		demo.LogCrit("pss get pubkey fail", "err", err)
	}

	// get the overlay addresses
	var leftBzzAddr string
	err = leftRPCClient.Call(&leftBzzAddr, "pss_baseAddr")
	if err != nil {
		demo.LogCrit("pss get baseaddr fail", "err", err)
	}
	var rightBzzAddr string
	err = rightRPCClient.Call(&rightBzzAddr, "pss_baseAddr")
	if err != nil {
		demo.LogCrit("pss get baseaddr fail", "err", err)
	}

	// make the nodes aware of each others' public keys
	err = leftRPCClient.Call(nil, "pss_setPeerPublicKey", rightPubKey, topic, rightBzzAddr)
	if err != nil {
		demo.LogCrit("pss set pubkey fail", "err", err)
	}
	err = rightRPCClient.Call(nil, "pss_setPeerPublicKey", leftPubKey, topic, leftBzzAddr)
	if err != nil {
		demo.LogCrit("pss set pubkey fail", "err", err)
	}

	// wait until the state of the swarm overlay network is ready
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = demo.WaitHealthy(ctx, 2, leftRPCClient, rightRPCClient)
	if err != nil {
		demo.LogCrit("health check fail", "err", err)
	}
	time.Sleep(time.Second) // because the healthy does not work

	// configure and start up pss client RPCs
	// we can use websockets ...
	leftClient, err := pssclient.NewClient(fmt.Sprintf("ws://%s", leftStack.WSEndpoint()))
	if err != nil {
		demo.LogCrit("pssclient 'left' create fail", "err", err)
	}
	// ... or unix sockets, the client handles both :)
	rightClient, err := pssclient.NewClient(rightStack.IPCEndpoint())
	if err != nil {
		demo.LogCrit("pssclient 'right' create fail", "err", err)
	}

	// set up generic ping protocol
	leftPing := pss.Ping{
		Pong: false,
		OutC: make(chan bool),
		InC:  make(chan bool),
	}
	leftProto := pss.NewPingProtocol(&leftPing)
	rightPing := pss.Ping{
		Pong: true,
		OutC: make(chan bool),
		InC:  make(chan bool),
	}
	rightProto := pss.NewPingProtocol(&rightPing)

	// run the pssclient protocols, wait for success
	// this registers the protocol handler in pss with topic generated from the protocol name and version
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	leftClient.RunProtocol(ctx, leftProto)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	rightClient.RunProtocol(ctx, rightProto)

	// add the 'right' peer
	leftClient.AddPssPeer(rightPubKey, common.FromHex(rightBzzAddr), pss.PingProtocol)

	time.Sleep(time.Second)

	// send ping
	leftPing.OutC <- false

	// get ping
	<-rightPing.InC
	demo.LogInfo("got ping")

	// get pong
	<-leftPing.InC
	demo.LogInfo("got pong")

	// the 'right' will receive the ping and send on the quit channel
	leftClient.Close()
	rightClient.Close()
}
