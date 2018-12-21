// pss send-to-self hello world
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/swarm"
	bzzapi "github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/pss"

	demo "github.com/tomochain/orderbook/common"
)

func newService(bzzdir string, bzzport int, bzznetworkid uint64) func(ctx *node.ServiceContext) (node.Service, error) {
	return func(ctx *node.ServiceContext) (node.Service, error) {

		// generate a new private key
		privkey, err := crypto.GenerateKey()
		if err != nil {
			demo.LogCrit("private key generate servicenode 'left' fail: %v")
		}

		// create necessary swarm params
		bzzconfig := bzzapi.NewConfig()
		bzzconfig.Path = bzzdir
		bzzconfig.Init(privkey)
		if err != nil {
			demo.LogCrit("unable to configure swarm", "err", err)
		}
		bzzconfig.Port = fmt.Sprintf("%d", bzzport)

		// shortcut to setting up a swarm node
		return swarm.NewSwarm(bzzconfig, nil)
	}
}

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
	// leftSvc := newService(leftStack.InstanceDir(), demo.BzzDefaultPort, demo.BzzDefaultNetworkId)
	err = leftStack.Register(leftSvc)
	if err != nil {
		demo.LogCrit("servicenode 'left' pss register fail", "err", err)
	}
	rightSvc := demo.NewSwarmService(rightStack, demo.BzzDefaultPort+1)
	// rightSvc := newService(rightStack.InstanceDir(), demo.BzzDefaultPort+1, demo.BzzDefaultNetworkId)
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
	// ... but the healthy functions doesnt seem to work, so we're stuck with timeout for now
	time.Sleep(time.Second)

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
	var rightBZZAddress string
	err = rightRPCClient.Call(&rightBZZAddress, "pss_baseAddr")
	if err != nil {
		demo.LogCrit("pss get pubkey fail", "err", err)
	}

	// get the receiver's public key
	var rightPublicKey string
	err = rightRPCClient.Call(&rightPublicKey, "pss_getPublicKey")
	if err != nil {
		demo.LogCrit("pss get pubkey fail", "err", err)
	}

	// make the sender aware of the receiver's public key
	err = leftRPCClient.Call(nil, "pss_setPeerPublicKey", rightPublicKey, topic, rightBZZAddress)
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
	demo.LogInfo("pss received", "msg", string(inmsg.Msg), "from", fmt.Sprintf("%x", inmsg.Key))

	// bring down the servicenodes
	sub.Unsubscribe()
	rightRPCClient.Close()
	leftRPCClient.Close()
	rightStack.Stop()
	leftStack.Stop()
}
