// send, receive, get notified about a message
package main

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"

	demo "github.com/tomochain/orderbook/common"
)

var (
	messageW = &sync.WaitGroup{}
)

// FooMsg : just a demo struct to send as message, it can be OrderBook, Action...
type FooMsg struct {
	Message string
}

// create a protocol that can take care of message sending
// the Run function is invoked upon connection
// it gets passed:
// * an instance of p2p.Peer, which represents the remote peer
// * an instance of p2p.MsgReadWriter, which is the io between the node and its peer

var (
	proto = p2p.Protocol{
		Name:    "foo",
		Version: 42,
		Length:  1,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {

			// simplest payload possible; a byte slice
			outmsg := &FooMsg{"foobar"}

			// send the message
			err := p2p.Send(rw, 0, outmsg)
			if err != nil {
				return fmt.Errorf("Send p2p message fail: %v", err)
			}
			demo.LogInfo("Sending message", "peer", p, "msg", outmsg)

			// wait for the message to come in from the other side
			// note that receive message event doesn't get emitted until we ReadMsg()
			inmsg, err := rw.ReadMsg()
			if err != nil {
				return fmt.Errorf("Receive p2p message fail: %v", err)
			}

			// need to define decoded Type, use reference so that no new object is created for the first time
			var val *FooMsg
			inmsg.Decode(&val)
			demo.LogInfo("Received message", "peer", p, "msg", val)

			// terminate the protocol
			return nil
		},
	}
)

func main() {

	// we need private keys for both servers
	privekey1, err := crypto.GenerateKey()
	if err != nil {
		demo.LogCrit("Generate private key #1 failed", "err", err)
	}
	privkey2, err := crypto.GenerateKey()
	if err != nil {
		demo.LogCrit("Generate private key #2 failed", "err", err)
	}

	// set up the two servers, unless map and list are reference, other must use reference, so that we can pass nil
	srv1 := demo.NewServer(privekey1, "foo", "42", &proto, 0)
	err = srv1.Start()
	if err != nil {
		demo.LogCrit("Start p2p.Server #1 failed", "err", err)
	}

	src2 := demo.NewServer(privkey2, "bar", "666", &proto, 31234)
	err = src2.Start()
	if err != nil {
		demo.LogCrit("Start p2p.Server #2 failed", "err", err)
	}

	// set up the event subscriptions on both servers
	// the Err() on the Subscription object returns when subscription is closed
	eventC1 := make(chan *p2p.PeerEvent)
	sub1 := srv1.SubscribeEvents(eventC1)
	messageW.Add(1)
	go func() {
		for {
			peerevent := <-eventC1
			if peerevent.Type == "add" {
				demo.LogDebug("Received peer add notification on node #1", "peer", peerevent.Peer)
			} else if peerevent.Type == "msgrecv" {
				demo.LogInfo("Received message nofification on node #1", "event", peerevent)

				messageW.Done()
				return
			}
		}
	}()

	eventC2 := make(chan *p2p.PeerEvent)
	sub2 := src2.SubscribeEvents(eventC2)
	messageW.Add(1)
	go func() {
		for {
			peerevent := <-eventC2
			if peerevent.Type == "add" {
				demo.LogDebug("Received peer add notification on node #2", "peer", peerevent.Peer)
			} else if peerevent.Type == "msgrecv" {
				demo.LogInfo("Received message nofification on node #2", "event", peerevent)
				messageW.Done()
				return
			}
		}
	}()

	// get the node instance of the second server
	mpde2 := src2.Self()

	// add it as a peer to the first node
	// the connection and crypto handshake will be performed automatically
	srv1.AddPeer(mpde2)

	// wait for each respective message to be delivered on both sides
	messageW.Wait()

	// terminate subscription loops and unsubscribe
	sub1.Unsubscribe()
	sub2.Unsubscribe()

	// stop the servers
	srv1.Stop()
	src2.Stop()
}
