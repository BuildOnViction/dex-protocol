// Previous "reply" example using p2p.protocols abstraction
package protocol

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/swarm/pss"
	demo "github.com/tomochain/orderbook/common"
)

var (
	messageW  = &sync.WaitGroup{}
	pssprotos = []*pss.Protocol{}
)

// FooMsg : struct message, using postal service over swarm with protocols which stores in leveldb
type FooMsg struct {
	V uint
}

// using the protocols abstraction, message structures are registered and their message codes handled automatically
var (
	fooProtocol = protocols.Spec{
		Name:       "Orderbook",
		Version:    42,
		MaxMsgSize: 1024,
		Messages: []interface{}{
			&FooMsg{},
		},
	}
	topic = pss.ProtocolTopic(&fooProtocol)
)

// the protocols abstraction enables use of an external handler function
type fooHandler struct {
	peer *p2p.Peer
}

// we will receive message in handle
func (self *fooHandler) handle(msg interface{}) error {
	foomsg, ok := msg.(*FooMsg)
	if !ok {
		return fmt.Errorf("invalid message : %v, peer: %v", msg, self.peer)
	}
	demo.LogInfo("received message", "foomsg", foomsg, "peer", self.peer)
	return nil
}

// create the protocol with the protocols extension
var (
	proto = p2p.Protocol{
		Name:    "foo",
		Version: 42,
		Length:  1,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			demo.LogWarn("running", "peer", p)
			// create the enhanced peer
			pp := protocols.NewPeer(p, rw, &fooProtocol)

			// send the message, then handle it to make sure protocol success
			go func() {
				outmsg := &FooMsg{
					V: 42,
				}
				err := pp.Send(outmsg)
				if err != nil {
					demo.LogError("Send p2p message fail", "err", err)
				}
				demo.LogInfo("sending message", "peer", p, "msg", outmsg)
			}()

			// protocols abstraction provides a separate blocking run loop for the peer
			// when this returns, the protocol will be terminated
			run := &fooHandler{
				peer: p,
			}
			err := pp.Run(run.handle)
			return err
		},
	}
)

func TestPssKey(t *testing.T) {
	rawurl := "enode://ce24c4f944a0a3614b691d839a6a89339d17abac3d69c0d24e806db45d1bdbe7afa53c02136e5ad952f43e6e7285cd3971e367d8789f4eb7306770f5af78755d@127.0.0.1:30101?discport=0"
	publicKey := "04ce24c4f944a0a3614b691d839a6a89339d17abac3d69c0d24e806db45d1bdbe7afa53c02136e5ad952f43e6e7285cd3971e367d8789f4eb7306770f5af78755d"
	bzzKey := "0x9984c9556ca87842c4ceb839518cd3648dc495d579f7af7f9ba49989bc207346"
	newNode, _ := discover.ParseNode(rawurl)
	pKey, _ := newNode.ID.Pubkey()
	pKeyStr := common.Bytes2Hex(crypto.FromECDSAPub(pKey))
	// pKey := "0x04" + newNode.ID.String()
	if publicKey != pKeyStr {
		t.Errorf("Want :%s, got %s", publicKey, pKeyStr)
	}

	// bzz is hash of full public key, not address (short of public key)
	pubkey := crypto.FromECDSAPub(pKey)
	expectedBzzKey := crypto.Keccak256Hash(pubkey).Hex()

	if expectedBzzKey != bzzKey {
		t.Errorf("Want :%s, got %s", bzzKey, expectedBzzKey)
	}
}
func Test2PeersCommunicationPss(t *testing.T) {

	// create two nodes
	privkey1, _ := crypto.HexToECDSA("3411b45169aa5a8312e51357db68621031020dcf46011d7431db1bbb6d3922ce")
	leftStack, err := demo.NewServiceNodeWithPrivateKey(privkey1, demo.P2pPort, 0, 0)
	if err != nil {
		demo.LogCrit(err.Error())
	}
	privkey2, _ := crypto.HexToECDSA("75c3e3150c0127af37e7e9df51430d36faa4c4660b6984c1edff254486d834e9")
	rightStack, err := demo.NewServiceNodeWithPrivateKey(privkey2, demo.P2pPort+1, 0, 0)
	if err != nil {
		demo.LogCrit(err.Error())
	}

	// p.Pss.Register(&topic, p.Handle)
	protocolSpecs := []*protocols.Spec{&fooProtocol}
	protocolArr := []*p2p.Protocol{&proto}

	// register the pss activated bzz services, using reference of slice so that we have modified list
	leftSvc := demo.NewSwarmServiceWithProtocolAndPrivateKey(leftStack, demo.BzzDefaultPort, protocolSpecs, protocolArr, &topic, &pssprotos, privkey1)
	err = leftStack.Register(leftSvc)
	if err != nil {
		demo.LogCrit("servicenode 'left' pss register fail", "err", err)
	}

	rightSvc := demo.NewSwarmServiceWithProtocolAndPrivateKey(rightStack, demo.BzzDefaultPort+1, protocolSpecs, protocolArr, &topic, &pssprotos, privkey2)
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

	// connect the nodes
	leftStack.Server().AddPeer(rightStack.Server().Self())

	// get the rpc clients
	leftRPCClient, err := leftStack.Attach()
	rightRPCClient, err := rightStack.Attach()

	// wait until the state of the swarm overlay network is ready
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	// wait with timeout for 2 nodes
	// err = demo.WaitHealthy(ctx, 2, leftRPCClient, rightRPCClient)
	// if err != nil {
	// 	demo.LogCrit("health check fail", "err", err)
	// }
	// time.Sleep(time.Second) // because the healthy does not work

	// get the overlay addresses
	var leftBzzAddr string
	err = rightRPCClient.Call(&leftBzzAddr, "pss_baseAddr")
	if err != nil {
		demo.LogCrit("pss get pubkey fail", "err", err)
	}
	var rightBzzAddr string
	err = leftRPCClient.Call(&rightBzzAddr, "pss_baseAddr")
	if err != nil {
		demo.LogCrit("pss get pubkey fail", "err", err)
	}

	demo.LogTrace("BaseAddr", "leftAddr", leftBzzAddr, "addrRight", rightBzzAddr)

	// get the publickeys
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

	// privkey, _ := crypto.GenerateKey()
	// rightPubKey = crypto.PubkeyToAddress(privkey.PublicKey).String()

	// set the peers' publickeys, so they will be mutual understanding
	err = leftRPCClient.Call(nil, "pss_setPeerPublicKey", rightPubKey, topic, rightBzzAddr)
	if err != nil {
		demo.LogCrit("pss set pubkey fail", "err", err)
	}
	err = rightRPCClient.Call(nil, "pss_setPeerPublicKey", leftPubKey, topic, leftBzzAddr)
	if err != nil {
		demo.LogCrit("pss set pubkey fail", "err", err)
	}

	// // set up the event subscriptions on both nodes
	// eventC1 := make(chan *p2p.PeerEvent)
	// sub1 := leftStack.Server().SubscribeEvents(eventC1)
	// messageW.Add(1)
	// go func() {
	// 	for {
	// 		select {
	// 		case peerevent := <-eventC1:
	// 			if peerevent.Type == "add" {
	// 				demo.LogDebug("Received peer add notification on node #1", "peer", peerevent.Peer)
	// 			} else if peerevent.Type == "msgrecv" {
	// 				demo.LogInfo("Received message nofification on node #1", "event", peerevent)
	// 				messageW.Done()
	// 			}
	// 		case <-sub1.Err():
	// 			return
	// 		}
	// 	}
	// }()

	// eventC2 := make(chan *p2p.PeerEvent)
	// sub2 := rightStack.Server().SubscribeEvents(eventC2)
	// messageW.Add(1)
	// go func() {
	// 	for {
	// 		select {
	// 		case peerevent := <-eventC2:
	// 			if peerevent.Type == "add" {
	// 				demo.LogDebug("Received peer add notification on node #2", "peer", peerevent.Peer)
	// 			} else if peerevent.Type == "msgrecv" {
	// 				demo.LogInfo("Received message nofification on node #2", "event", peerevent)
	// 				messageW.Done()
	// 			}
	// 		case <-sub2.Err():
	// 			return
	// 		}
	// 	}
	// }()

	// // wait for each respective message to be delivered on both sides
	// messageW.Wait()

	msgC := make(chan pss.APIMsg)
	sub, err := rightRPCClient.Subscribe(context.Background(), "pss", msgC, "receive", topic)
	// last one

	// get the incoming message
	go func() {
		ciphertext := hex.EncodeToString([]byte("hehe"))
		err = leftRPCClient.Call(nil, "pss_sendRaw", rightBzzAddr, topic, ciphertext)
		if err != nil {
			demo.LogCrit("Pss send fail", "err", err)
		}
	}()

	inmsg := <-msgC
	fmt.Println("In message", inmsg.Msg)

	sub.Unsubscribe()
	// terminate subscription loops and unsubscribe
	// sub1.Unsubscribe()
	// sub2.Unsubscribe()
	rightRPCClient.Close()
	leftRPCClient.Close()
	rightStack.Stop()
	leftStack.Stop()
}
