// Node stack with ping/pong and API reporting
package protocol

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/rpc"

	demo "github.com/tomochain/orderbook/common"
)

var (
	stackW = &sync.WaitGroup{}
)

type FooPingMsg struct {
	Pong    bool
	Created time.Time
}

// the service we want to offer on the node
// it must implement the node.Service interface
type fooService struct {
	pongcount int
	pingC     map[discover.NodeID]chan struct{}
}

// specify API structs that carry the methods we want to use
func (service *fooService) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "foo",
			Version:   "42",
			Service: &FooAPI{
				running:   true,
				pongcount: &service.pongcount,
				pingC:     service.pingC,
			},
			Public: true,
		},
	}
}

// the p2p.Protocol to run
// sends a ping to its peer, waits pong
func (service *fooService) Protocols() []p2p.Protocol {

	return []p2p.Protocol{
		{
			Name:    "fooping",
			Version: 666,
			Length:  1,
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				// create the channel when a connection is made
				demo.LogInfo("Protocol method run")
				service.pingC[p.ID()] = make(chan struct{})
				pingcount := 0

				// create the message structure

				// we don't know if we're awaiting anything at the time of the kill so this subroutine will run till the application ends
				go func() {
					for {
						// listen for new message
						msg, err := rw.ReadMsg()
						if err != nil {
							demo.LogWarn("Receive p2p message fail", "err", err)
							break
						}

						// decode the message and check the contents
						var decodedmsg FooPingMsg
						err = msg.Decode(&decodedmsg)
						if err != nil {
							demo.LogError("Decode p2p message fail", "err", err)
							break
						}

						// if we get a pong, update our pong counter
						// if not, send pong
						if decodedmsg.Pong {
							service.pongcount++
							demo.LogDebug("received pong", "peer", p, "count", service.pongcount)
						} else {
							demo.LogDebug("received ping", "peer", p)
							pingmsg := &FooPingMsg{
								Pong:    true,
								Created: time.Now(),
							}
							err := p2p.Send(rw, 0, pingmsg)
							if err != nil {
								demo.LogError("Send p2p message fail", "err", err)
								break
							}
							demo.LogDebug("sent pong", "peer", p)
						}
					}
				}()

				// pings are invoked through the API using a channel
				// when this channel is closed we quit the protocol
				for {
					// wait for signal to send ping
					_, ok := <-service.pingC[p.ID()]
					if !ok {
						demo.LogDebug("break protocol", "peer", p)
						break
					}

					// send ping
					pingmsg := &FooPingMsg{
						Pong:    false,
						Created: time.Now(),
					}

					// either handler or sender should be asynchronous, otherwise we might deadlock
					go p2p.Send(rw, 0, pingmsg)
					pingcount++
					demo.LogInfo("sent ping", "peer", p, "count", pingcount)
				}

				return nil
			},
		},
	}
}

func (service *fooService) Start(srv *p2p.Server) error {
	return nil
}

func (service *fooService) Stop() error {
	return nil
}

// Specify the API
// in this example we don't care about who the pongs comes from, we count them all
// note it is a bit fragile; we don't check for closed channels
type FooAPI struct {
	running   bool
	pongcount *int
	pingC     map[discover.NodeID]chan struct{}
}

func (api *FooAPI) Increment() {
	*api.pongcount++
}

// invoke a single ping
func (api *FooAPI) Ping(id discover.NodeID) error {
	if api.running {
		api.pingC[id] <- struct{}{}
	}
	return nil
}

// quit the ping protocol
func (api *FooAPI) Quit(id discover.NodeID) error {
	demo.LogDebug("quitting API", "peer", id)
	if api.pingC[id] == nil {
		return fmt.Errorf("unknown peer")
	}
	api.running = false
	close(api.pingC[id])
	return nil
}

// return the amounts of pongs received
func (api *FooAPI) PongCount() (int, error) {
	return *api.pongcount, nil
}

func TestPss(t *testing.T) {
	rawurl := "enode://ce24c4f944a0a3614b691d839a6a89339d17abac3d69c0d24e806db45d1bdbe7afa53c02136e5ad952f43e6e7285cd3971e367d8789f4eb7306770f5af78755d@127.0.0.1:30101?discport=0"
	publicKey := "0x04ce24c4f944a0a3614b691d839a6a89339d17abac3d69c0d24e806db45d1bdbe7afa53c02136e5ad952f43e6e7285cd3971e367d8789f4eb7306770f5af78755d"
	newNode, _ := discover.ParseNode(rawurl)
	pKey := "0x04" + newNode.ID.String()

	t.Logf("Node ID :%t", publicKey == pKey)
}

func Test2PeersCommunication(t *testing.T) {

	// create the two nodes
	stack1, err := demo.NewServiceNode(demo.P2pPort, 0, 0)
	if err != nil {
		demo.LogCrit("Create servicenode #1 fail", "err", err)
	}
	stack2, err := demo.NewServiceNode(demo.P2pPort+1, 0, 0)
	if err != nil {
		demo.LogCrit("Create servicenode #2 fail", "err", err)
	}

	// wrapper function for servicenode to start the service
	foosvc := func(ctx *node.ServiceContext) (node.Service, error) {
		return &fooService{
			pingC: make(map[discover.NodeID]chan struct{}),
		}, nil
	}

	// register adds the service to the services the servicenode starts when started
	err = stack1.Register(foosvc)
	if err != nil {
		demo.LogCrit("Register service in servicenode #1 fail", "err", err)
	}
	err = stack2.Register(foosvc)
	if err != nil {
		demo.LogCrit("Register service in servicenode #2 fail", "err", err)
	}

	// start the nodes
	err = stack1.Start()
	if err != nil {
		demo.LogCrit("servicenode #1 start failed", "err", err)
	}
	err = stack2.Start()
	if err != nil {
		demo.LogCrit("servicenode #2 start failed", "err", err)
	}

	// connect to the servicenode RPCs
	rpcclient1, err := rpc.Dial(filepath.Join(stack1.DataDir(), demo.IPCName))
	if err != nil {
		demo.LogCrit("connect to servicenode #1 IPC fail", "err", err)
	}
	defer os.RemoveAll(stack1.DataDir())

	rpcclient2, err := rpc.Dial(filepath.Join(stack2.DataDir(), demo.IPCName))
	if err != nil {
		demo.LogCrit("connect to servicenode #2 IPC fail", "err", err)
	}
	defer os.RemoveAll(stack2.DataDir())

	// display that the initial pong counts are 0
	var count int
	err = rpcclient1.Call(&count, "foo_pongCount")
	if err != nil {
		demo.LogCrit("servicenode #1 pongcount RPC failed", "err", err)
	}
	demo.LogInfo("servicenode #1 before ping", "pongcount", count)

	err = rpcclient2.Call(&count, "foo_pongCount")
	if err != nil {
		demo.LogCrit("servicenode #2 pongcount RPC failed", "err", err)
	}
	demo.LogInfo("servicenode #2 before ping", "pongcount", count)

	// get the server instances
	srv1 := stack1.Server()
	srv2 := stack2.Server()

	// subscribe to peerevents
	eventC1 := make(chan *p2p.PeerEvent)
	sub1 := srv1.SubscribeEvents(eventC1)

	eventC2 := make(chan *p2p.PeerEvent)
	sub2 := srv2.SubscribeEvents(eventC2)

	// connect the nodes
	p2pNode2 := srv2.Self()
	srv1.AddPeer(p2pNode2)

	// fork and do the pinging
	stackW.Add(2)
	pingmax1 := 4
	pingmax2 := 2

	go func() {
		// when we get the add event, we know we are connected
		ev := <-eventC1
		if ev.Type != "add" {
			demo.LogError("server #1 expected peer add", "eventtype", ev.Type)
			stackW.Done()
			return
		}
		demo.LogDebug("server #1 connected", "peer", ev.Peer)

		// send the pings
		for i := 0; i < pingmax1; i++ {
			err := rpcclient1.Call(nil, "foo_ping", ev.Peer)
			if err != nil {
				demo.LogError("server #1 RPC ping fail", "err", err)
				stackW.Done()
				break
			}
		}

		// wait for all msgrecv events
		// pings we receive, and pongs we expect from pings we sent
		for i := 0; i < pingmax2+pingmax1; {
			ev := <-eventC1
			demo.LogWarn("msg", "type", ev.Type, "i", i)
			if ev.Type == "msgrecv" {
				i++
			}
		}

		stackW.Done()
	}()

	// mirrors the previous go func
	go func() {
		ev := <-eventC2
		if ev.Type != "add" {
			demo.LogError("expected peer add", "eventtype", ev.Type)
			stackW.Done()
			return
		}
		demo.LogDebug("server #2 connected", "peer", ev.Peer)
		for i := 0; i < pingmax2; i++ {
			err := rpcclient2.Call(nil, "foo_ping", ev.Peer)
			if err != nil {
				demo.LogError("server #2 RPC ping fail", "err", err)
				stackW.Done()
				break
			}
		}

		for i := 0; i < pingmax1+pingmax2; {
			ev := <-eventC2
			if ev.Type == "msgrecv" {
				demo.LogWarn("msg", "type", ev.Type, "i", i)
				i++
			}
		}

		stackW.Done()
	}()

	// wait for the two ping pong exchanges to finish
	stackW.Wait()

	// tell the API to shut down
	// this will disconnect the peers and close the channels connecting API and protocol
	err = rpcclient1.Call(nil, "foo_quit", p2pNode2.ID.String())
	if err != nil {
		demo.LogError("server #1 RPC quit fail", "err", err)
	}
	err = rpcclient2.Call(nil, "foo_quit", srv1.Self().ID.String())
	if err != nil {
		demo.LogError("server #2 RPC quit fail", "err", err)
	}

	// disconnect will generate drop events
	for {
		ev := <-eventC1
		if ev.Type == "drop" {
			break
		}
	}
	for {
		ev := <-eventC2
		if ev.Type == "drop" {
			break
		}
	}

	// proudly inspect the results
	err = rpcclient1.Call(&count, "foo_pongCount")
	if err != nil {
		demo.LogCrit("servicenode #1 pongcount RPC failed", "err", err)
	}
	demo.LogInfo("servicenode #1 after ping", "pongcount", count)

	err = rpcclient2.Call(&count, "foo_pongCount")
	if err != nil {
		demo.LogCrit("servicenode #2 pongcount RPC failed", "err", err)
	}
	demo.LogInfo("servicenode #2 after ping", "pongcount", count)

	// bring down the servicenodes after receive 4 pongcount at server 1 and 2 pongcount at server 2
	sub1.Unsubscribe()
	sub2.Unsubscribe()
	stack1.Stop()
	stack2.Stop()
}
