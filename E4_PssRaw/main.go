// pss send message using external encryption
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/ethereum/go-ethereum/swarm/pss"

	demo "github.com/tomochain/orderbook/common"
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
	err = leftRPCClient.Call(&leftBzzAddr, "pss_baseAddr")
	if err != nil {
		demo.LogCrit("pss get baseaddr fail", "err", err)
	}
	var rightBzzAddr string
	err = rightRPCClient.Call(&rightBzzAddr, "pss_baseAddr")
	if err != nil {
		demo.LogCrit("pss get baseaddr fail", "err", err)
	}

	// generate the encryption key to use and encrypt the message with it
	rightExternalKey, err := ecies.GenerateKey(rand.Reader, crypto.S256(), nil)
	if err != nil {
		demo.LogCrit("generate external encryption key fail", "err", err)
	}
	m := []byte("xyzzy")
	ciphertext, err := ecies.Encrypt(rand.Reader, &rightExternalKey.PublicKey, m, nil, nil)
	if err != nil {
		demo.LogCrit("external message encryption fail", "err", err)
	}

	// send message using symmetric encryption
	// since it's sent to ourselves, it will not go through pss forwarding
	err = leftRPCClient.Call(nil, "pss_sendRaw", rightBzzAddr, topic, ciphertext)
	if err != nil {
		demo.LogCrit("Pss send fail", "err", err)
	}

	// get the incoming message
	inmsg := <-msgC

	// decrypt the message
	plaintext, err := rightExternalKey.Decrypt(inmsg.Msg, nil, nil)
	demo.LogInfo("Pss received", "msg", string(plaintext), "from", fmt.Sprintf("%x", inmsg.Key))

	// bring down the servicenodes
	sub.Unsubscribe()
	rightRPCClient.Close()
	leftRPCClient.Close()
	rightStack.Stop()
	leftStack.Stop()
}
