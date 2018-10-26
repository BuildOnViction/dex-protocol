// pss send-to-self hello world
package main

import (
	"context"
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

	// subscribe to incoming messages on the receiving sevicenode
	// this will register a message handler on the specified topic
	msgC := make(chan pss.APIMsg)
	sub, err := rightRPCClient.Subscribe(context.Background(), "pss", msgC, "receive", topic)

	// supply no address for routing
	rightBzzaddr := "0x"

	// get the receiver's public key
	var rightPublicKey string
	err = rightRPCClient.Call(&rightPublicKey, "pss_getPublicKey")
	if err != nil {
		demo.LogCrit("pss get pubkey fail", "err", err)
	}

	// make the sender aware of the receiver's public key
	err = leftRPCClient.Call(nil, "pss_setPeerPublicKey", rightPublicKey, topic, rightBzzaddr)
	if err != nil {
		demo.LogCrit("pss get pubkey fail", "err", err)
	}

	// send message using asymmetric encryption
	// since it's sent to ourselves, it will not go through pss forwarding
	err = leftRPCClient.Call(nil, "pss_sendAsym", rightPublicKey, topic, common.ToHex([]byte("bar")))
	if err != nil {
		demo.LogCrit("pss send fail", "err", err)
	}

	// get the incoming message
	inmsg := <-msgC
	demo.LogInfo("Pss received", "msg", string(inmsg.Msg), "from", inmsg.Key)

	// bring down the servicenodes
	sub.Unsubscribe()
	rightRPCClient.Close()
	leftRPCClient.Close()
	centerStack.Stop()
	rightStack.Stop()
	leftStack.Stop()
}
