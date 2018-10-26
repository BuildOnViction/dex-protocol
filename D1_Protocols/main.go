// Previous "reply" example using p2p.protocols abstraction
package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"

	demo "../common"
)

var (
	messageW = &sync.WaitGroup{}
)

type FooMsg struct {
	V uint
}

// using the protocols abstraction, message structures are registered and their message codes handled automatically
var (
	fooProtocol = protocols.Spec{
		Name:       demo.FooProtocolName,
		Version:    demo.FooProtocolVersion,
		MaxMsgSize: demo.FooProtocolMaxMsgSize,
		Messages: []interface{}{
			&FooMsg{},
		},
	}
)

// the protocols abstraction enables use of an external handler function
type fooHandler struct {
	peer *p2p.Peer
}

func (self *fooHandler) handle(ctx context.Context, msg interface{}) error {
	foomsg, ok := msg.(*FooMsg)
	if !ok {
		return fmt.Errorf("invalid message", "msg", msg, "peer", self.peer)
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
			demo.LogInfo("Protocol method run")
			// create the enhanced peer
			pp := protocols.NewPeer(p, rw, &fooProtocol)

			// send the message
			outmsg := &FooMsg{
				V: 42,
			}
			ctx := context.TODO()
			err := pp.Send(ctx, outmsg)
			if err != nil {
				demo.LogError("Send p2p message fail", "err", err)
			}
			demo.LogInfo("sending message", "peer", p, "msg", outmsg)

			// protocols abstraction provides a separate blocking run loop for the peer
			// a separate handler function is passed to that loop to process incoming messages
			run := &fooHandler{
				peer: p,
			}
			err = pp.Run(run.handle)

			// terminate the protocol
			return err
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
	eventC1 := make(chan *p2p.PeerEvent)
	sub1 := srv1.SubscribeEvents(eventC1)
	messageW.Add(1)
	go func() {
		for {
			select {
			case peerevent := <-eventC1:
				if peerevent.Type == "add" {
					demo.LogDebug("Received peer add notification on node #1", "peer", peerevent.Peer)
				} else if peerevent.Type == "msgrecv" {
					demo.LogInfo("Received message nofification on node #1", "event", peerevent)
					messageW.Done()
				}
			case <-sub1.Err():
				return
			}
		}
	}()

	eventC2 := make(chan *p2p.PeerEvent)
	sub2 := srv2.SubscribeEvents(eventC2)
	messageW.Add(1)
	go func() {
		for {
			select {
			case peerevent := <-eventC2:
				if peerevent.Type == "add" {
					demo.LogDebug("Received peer add notification on node #2", "peer", peerevent.Peer)
				} else if peerevent.Type == "msgrecv" {
					demo.LogInfo("Received message nofification on node #2", "event", peerevent)
					messageW.Done()
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
	messageW.Wait()

	// terminate subscription loops and unsubscribe
	sub1.Unsubscribe()
	sub2.Unsubscribe()

	// stop the servers
	srv1.Stop()
	srv2.Stop()
}
