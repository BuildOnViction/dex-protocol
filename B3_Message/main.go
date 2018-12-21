// trigger p2p message with RPC
package main

import (
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	demo "github.com/tomochain/orderbook/common"
)

var (
	protoW   = &sync.WaitGroup{}
	messageW = &sync.WaitGroup{}
	msgC     = make(chan string)
	ipcpath  = ".demo.ipc"
)

// create a protocol that can take care of message sending
// the Run function is invoked upon connection
// it gets passed:
// * an instance of p2p.Peer, which represents the remote peer
// * an instance of p2p.MsgReadWriter, which is the io between the node and its peer

// FooMsg : struct message
type FooMsg struct {
	Content string
}

var (
	proto = p2p.Protocol{
		Name:    "foo",
		Version: 42,
		Length:  1,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {

			// only one of the peers will send this
			content, ok := <-msgC
			if ok {
				outmsg := &FooMsg{
					Content: content,
				}

				// send the message
				err := p2p.Send(rw, 0, outmsg)
				if err != nil {
					return fmt.Errorf("Send p2p message fail: %v", err)
				}
				demo.LogInfo("Sending message", "peer", p, "msg", outmsg)
			}

			// wait for the subscriptions to end
			messageW.Wait()
			protoW.Done()

			// terminate the protocol
			return nil
		},
	}
)

// FooAPI : service
type FooAPI struct {
	sent bool
}

func (api *FooAPI) SendMsg(content string) error {
	if api.sent {
		return fmt.Errorf("Already sent")
	}
	msgC <- content
	close(msgC)
	api.sent = true
	return nil
}

func newRPCServer() (*rpc.Server, error) {
	// set up the RPC server
	rpcsrv := rpc.NewServer()
	err := rpcsrv.RegisterName("foo", &FooAPI{})
	if err != nil {
		return nil, fmt.Errorf("Register API method(s) fail: %v", err)
	}

	// create IPC endpoint
	ipclistener, err := net.Listen("unix", ipcpath)
	if err != nil {
		return nil, fmt.Errorf("IPC endpoint create fail: %v", err)
	}

	// mount RPC server on IPC endpoint
	// it will automatically detect and serve any valid methods
	go func() {
		err = rpcsrv.ServeListener(ipclistener)
		if err != nil {
			demo.LogCrit("Mount RPC on IPC fail", "err", err)
		}
	}()

	return rpcsrv, nil
}

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
	go func() {
		for {
			select {
			case peerEvent := <-eventC1:
				// add peer
				if peerEvent.Type == "add" {
					demo.LogDebug("Received peer add notification on node #1", "peer", peerEvent.Peer)
				} else if peerEvent.Type == "msgsend" {
					// send message
					demo.LogInfo("Received message send notification on node #1", "event", peerEvent)
					messageW.Done()
				}
			case <-sub1.Err():
				return
			}
		}
	}()

	eventC2 := make(chan *p2p.PeerEvent)
	sub2 := srv2.SubscribeEvents(eventC2)
	go func() {
		for {
			select {
			case peerEvent := <-eventC2:
				// the same for server 2
				if peerEvent.Type == "add" {
					demo.LogDebug("Received peer add notification on node #2", "peer", peerEvent.Peer)
				} else if peerEvent.Type == "msgsend" {
					demo.LogInfo("Received message send notification on node #2", "event", peerEvent)
					messageW.Done()
				}
			case <-sub2.Err():
				return
			}
		}
	}()

	// create and start RPC server
	rpcsrv, err := newRPCServer()
	if err != nil {
		demo.LogCrit(err.Error())
	}
	defer os.Remove(ipcpath)

	// get the node instance of the second server
	node2 := srv2.Self()

	// add it as a peer to the first node
	// the connection and crypto handshake will be performed automatically
	srv1.AddPeer(node2)

	// create an IPC client, then trigger rpc call to send message from node to node
	rpcclient, err := rpc.Dial(ipcpath)
	if err != nil {
		demo.LogCrit("IPC dial fail", "err", err)
	}

	// wait for one message be sent, and both protocols to end
	messageW.Add(1)
	protoW.Add(2)

	// call the RPC method
	err = rpcclient.Call(nil, "foo_sendMsg", "foobar")
	if err != nil {
		demo.LogCrit("RPC call fail", "err", err)
	}

	// wait for protocols to finish
	protoW.Wait()

	// terminate subscription loops and unsubscribe
	sub1.Unsubscribe()
	sub2.Unsubscribe()

	// stop the servers
	rpcsrv.Stop()
	srv1.Stop()
	srv2.Stop()
}
