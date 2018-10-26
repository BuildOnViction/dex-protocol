// pss send-to-self hello world
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/swarm/pss"

	demo "../common"
)

func main() {

	// create three nodes
	leftStack, err := demo.NewServiceNode(demo.P2pPort, 0, 0)
	if err != nil {
		demo.LogCrit(err.Error())
	}
	rightStack, err := demo.NewServiceNode(demo.P2pPort+1, 0, 0)
	if err != nil {
		demo.LogCrit(err.Error())
	}
	centerStack, err := demo.NewServiceNode(demo.P2pPort+2, 0, 0)
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
	centerSvc := demo.NewSwarmService(centerStack, demo.BzzDefaultPort+2)
	err = centerStack.Register(centerSvc)
	if err != nil {
		demo.LogCrit("servicenode 'middle' pss register fail", "err", err)
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
	err = centerStack.Start()
	if err != nil {
		demo.LogCrit("servicenode start failed", "err", err)
	}
	defer os.RemoveAll(centerStack.DataDir())

	// connect the nodes to the middle
	centerStack.Server().AddPeer(leftStack.Server().Self())
	centerStack.Server().AddPeer(rightStack.Server().Self())

	// get the rpc clients
	leftRPCClient, err := leftStack.Attach()
	rightRPCClient, err := rightStack.Attach()

	// wait until the state of the swarm overlay network is ready
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = demo.WaitHealthy(ctx, 2, leftRPCClient, rightRPCClient)
	if err != nil {
		demo.LogCrit("health check fail", "err", err)
	}
	time.Sleep(time.Second) // because the healthy does not work

	// get a valid topic byte
	var topic string
	err = leftRPCClient.Call(&topic, "pss_stringToTopic", "foo")
	if err != nil {
		demo.LogCrit("pss string to topic fail", "err", err)
	}

	// subscribe to incoming messages on both servicenodes
	// this will register message handlers, needed to receive reciprocal comms
	leftMsgC := make(chan pss.APIMsg)
	leftSubPSS, err := leftRPCClient.Subscribe(context.Background(), "pss", leftMsgC, "receive", topic)
	if err != nil {
		demo.LogCrit("pss subscribe error", "err", err)
	}
	rightMsgC := make(chan pss.APIMsg)
	rightSubPSS, err := rightRPCClient.Subscribe(context.Background(), "pss", rightMsgC, "receive", topic)
	if err != nil {
		demo.LogCrit("pss subscribe error", "err", err)
	}

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

	// activate handshake on both sides
	err = leftRPCClient.Call(nil, "pss_addHandshake", topic)
	if err != nil {
		demo.LogCrit("pss handshake activate fail", "err", err)
	}
	err = rightRPCClient.Call(nil, "pss_addHandshake", topic)
	if err != nil {
		demo.LogCrit("pss handshake activate fail", "err", err)
	}

	// initiate handshake and retrieve symkeys
	var symkeyids []string
	err = leftRPCClient.Call(&symkeyids, "pss_handshake", rightPubKey, topic, true, true)
	if err != nil {
		demo.LogCrit("handshake fail", "err", err)
	}

	// convert the pubkey to hex string
	// send message using asymmetric encryption, the message will be in hexa format, or we can use sendRaw with custom encryption engine
	err = leftRPCClient.Call(nil, "pss_sendSym", symkeyids[0], topic, common.ToHex([]byte("bar")))
	if err != nil {
		demo.LogCrit("pss send fail", "err", err)
	}

	// get the incoming message with custom channel
	for {
		inmsg := <-rightMsgC
		if !inmsg.Asymmetric {
			// string init from bytes, default it is hex string to represent bytes
			demo.LogInfo("Pss received", "msg", string(inmsg.Msg), "from", fmt.Sprintf("%x", inmsg.Key))
			break
		}
	}

	// bring down the servicenodes
	leftSubPSS.Unsubscribe()
	rightSubPSS.Unsubscribe()
	rightRPCClient.Close()
	leftRPCClient.Close()
	centerStack.Stop()
	rightStack.Stop()
	leftStack.Stop()
}
