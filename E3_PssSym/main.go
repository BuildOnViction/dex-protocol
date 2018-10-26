// pss send symmetrically encrypted message
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/swarm/pss"

	demo "../common"
)

func main() {

	// create two nodes
	leftStack, err := demo.NewServiceNode(demo.P2pPort, 0, 0)
	if err != nil {
		demo.LogCrit(err.Error())
	}
	rightStack, err := demo.NewServiceNode(demo.P2pPort+1, 0, 0)
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

	// get the recipient node's swarm overlay address
	var leftBzzAddr string
	err = rightRPCClient.Call(&leftBzzAddr, "pss_baseAddr")
	if err != nil {
		demo.LogCrit("pss get baseaddr fail", "err", err)
	}
	var rightBzzAddr string
	err = rightRPCClient.Call(&rightBzzAddr, "pss_baseAddr")
	if err != nil {
		demo.LogCrit("pss get baseaddr fail", "err", err)
	}

	symkey := make([]byte, 32)
	c, err := rand.Read(symkey)
	if err != nil {
		demo.LogCrit("symkey gen fail", "err", err)
	} else if c < 32 {
		demo.LogCrit("symkey size mismatch, expected 32", "size", c)
	}

	var leftSymKeyID string
	err = leftRPCClient.Call(&leftSymKeyID, "pss_setSymmetricKey", symkey, topic, rightBzzAddr, true)
	if err != nil {
		demo.LogCrit("Pss set symkey fail", "err", err)
	}

	var rightSymKeyID string
	err = rightRPCClient.Call(&rightSymKeyID, "pss_setSymmetricKey", symkey, topic, leftBzzAddr, true)
	if err != nil {
		demo.LogCrit("Pss set symkey fail", "err", err)
	}

	// send message using symmetric encryption
	// since it's sent to ourselves, it will not go through pss forwarding
	err = leftRPCClient.Call(nil, "pss_sendSym", leftSymKeyID, topic, common.ToHex([]byte("bar")))
	if err != nil {
		demo.LogCrit("Pss send fail", "err", err)
	}

	// get the incoming message
	inmsg := <-msgC
	demo.LogInfo("Pss received", "msg", string(inmsg.Msg), "from", fmt.Sprintf("%x", inmsg.Key))

	// bring down the servicenodes
	sub.Unsubscribe()
	rightRPCClient.Close()
	leftRPCClient.Close()
	rightStack.Stop()
	leftStack.Stop()
}
