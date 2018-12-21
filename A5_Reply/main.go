// send, receive, get notified about a message
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"

	demo "github.com/tomochain/orderbook/common"
)

var (
	protoW = &sync.WaitGroup{}
	pingW  = &sync.WaitGroup{}
)

type FooPingMsg struct {
	Pong    bool
	Created time.Time
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

			pingW.Add(1)
			ponged := false

			// create the message structure
			msg := FooPingMsg{
				Pong:    false,
				Created: time.Now(),
			}

			// send the message
			err := p2p.Send(rw, 0, msg)
			if err != nil {
				return fmt.Errorf("Send p2p message fail: %v", err)
			}
			demo.LogInfo("sending ping", "peer", p)

			for !ponged {
				// wait for the message to come in from the other side
				// note that receive message event doesn't get emitted until we ReadMsg()
				msg, err := rw.ReadMsg()
				if err != nil {
					return fmt.Errorf("Receive p2p message fail: %v", err)
				}

				// decode the message and check the contents
				var decodedmsg FooPingMsg
				err = msg.Decode(&decodedmsg)
				if err != nil {
					return fmt.Errorf("Decode p2p message fail: %v", err)
				}

				if decodedmsg.Pong {
					demo.LogInfo("received pong", "peer", p)
					ponged = true
					pingW.Done()
				} else {
					demo.LogInfo("received ping", "peer", p)
					msg := FooPingMsg{
						Pong:    true,
						Created: time.Now(),
					}
					err := p2p.Send(rw, 0, msg)
					if err != nil {
						return fmt.Errorf("Send p2p message fail: %v", err)
					}
					demo.LogInfo("sent pong", "peer", p)
				}

			}

			// terminate the protocol after all involved have completed the cycle
			pingW.Wait()
			protoW.Done()
			return nil
		},
	}
)

func main() {

	// we need private keys for both servers
	privkey1, err := crypto.GenerateKey()
	if err != nil {
		demo.LogCrit("Generate private key #1 failed", "err", err)
	}
	privkey2, err := crypto.GenerateKey()
	if err != nil {
		demo.LogCrit("Generate private key #2 failed", "err", err)
	}

	// set up the two servers
	srv1 := demo.NewServer(privkey1, "foo", "42", &proto, 0)
	err = srv1.Start()
	if err != nil {
		demo.LogCrit("Start p2p.Server #1 failed", "err", err)
	}

	srv2 := demo.NewServer(privkey2, "bar", "666", &proto, 31234)
	err = srv2.Start()
	if err != nil {
		demo.LogCrit("Start p2p.Server #2 failed", "err", err)
	}

	// set up the event subscriptions on both servers
	// the Err() on the Subscription object returns when subscription is closed
	eventC1 := make(chan *p2p.PeerEvent)
	sub1 := srv1.SubscribeEvents(eventC1)
	protoW.Add(1)
	go func() {
		for {
			select {
			case peerevent := <-eventC1:
				if peerevent.Type == "add" {
					demo.LogDebug("Received peer add notification on node #1", "peer", peerevent.Peer)
				} else if peerevent.Type == "msgrecv" {
					demo.LogInfo("Received message nofification on node #1", "event", peerevent)
				}
			case <-sub1.Err():
				return
			}
		}
	}()

	eventC2 := make(chan *p2p.PeerEvent)
	sub2 := srv2.SubscribeEvents(eventC2)
	protoW.Add(1)
	go func() {
		for {
			select {
			case peerevent := <-eventC2:
				if peerevent.Type == "add" {
					demo.LogDebug("Received peer add notification on node #2", "peer", peerevent.Peer)
				} else if peerevent.Type == "msgrecv" {
					demo.LogInfo("Received message nofification on node #2", "event", peerevent)
				}
			case <-sub2.Err():
				return
			}
		}
	}()

	// get the node instance of the second server
	node2 := srv2.Self()

	// add it as a peer to the first node
	// the connection and crypto handshake will be performed automatically
	srv1.AddPeer(node2)

	// wait for each respective message to be delivered on both sides
	protoW.Wait()

	// terminate subscription loops and unsubscribe
	sub1.Unsubscribe()
	sub2.Unsubscribe()

	// stop the servers
	srv1.Stop()
	srv2.Stop()
}
