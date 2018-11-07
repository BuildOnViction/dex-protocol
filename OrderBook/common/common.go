package common

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/fatih/color"

	// peer.ID = enode.ID, NodeID = enode.ID
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm"
	bzzapi "github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/network"

	// we may need to re-implement this
	"github.com/ethereum/go-ethereum/swarm/pss"
)

var (
	// custom log, easily grep'able
	Log = log.New("DemoLog", "v1.0.1")

	// our working directory
	BasePath string

	// self-explanatory command line arguments
	verbose = flag.Bool("v", false, "more verbose logs")
)

// setup logging
// set up remote node, if present
func init() {
	var err error

	flag.Parse()

	// get the working directory
	BasePath, err = os.Getwd()
	if err != nil {
		Log.Crit("Could not determine working directory", "err", err)
	}

	// ensure good log formats for terminal
	// handle verbosity flag
	hs := log.StreamHandler(os.Stderr, log.TerminalFormat(true))
	loglevel := log.LvlInfo
	if *verbose {
		loglevel = log.LvlTrace
	}
	hf := log.LvlFilterHandler(loglevel, hs)
	h := log.CallerFileHandler(hf)
	log.Root().SetHandler(h)
}

// NewSwarmService : short cut for create swarm service
func NewSwarmService(stack *node.Node, bzzport int) func(ctx *node.ServiceContext) (node.Service, error) {
	return NewSwarmServiceWithProtocol(stack, bzzport, nil, nil, nil, nil)
}

func setInterval(someFunc func(), milliseconds int, async bool) chan bool {

	// How often to fire the passed in function
	// in milliseconds
	interval := time.Duration(milliseconds) * time.Millisecond

	// Setup the ticket and the channel to signal
	// the ending of the interval
	ticker := time.NewTicker(interval)
	clear := make(chan bool)

	// Put the selection in a go routine
	// so that the for loop is none blocking
	go func() {
		for {

			select {
			case <-ticker.C:
				if async {
					// This won't block
					go someFunc()
				} else {
					// This will block
					someFunc()
				}
			case <-clear:
				ticker.Stop()
				return
			}

		}
	}()

	// We return the channel so we can pass in
	// a value to it to clear the interval
	return clear

}

func NewSwarmServiceWithProtocolAndPrivateKey(stack *node.Node, bzzport int, specs []*protocols.Spec, protocols []*p2p.Protocol, topic *pss.Topic, pssprotos *[]*pss.Protocol, prvkey *ecdsa.PrivateKey) func(ctx *node.ServiceContext) (node.Service, error) {
	return func(ctx *node.ServiceContext) (node.Service, error) {

		// create the swarm overlay address
		//chbookaddr := crypto.PubkeyToAddress(prvkey.PublicKey)

		// configure and create a swarm instance
		bzzdir := stack.InstanceDir() // todo: what is the difference between this and datadir?

		bzzconfig := bzzapi.NewConfig()
		bzzconfig.NetworkID = BzzDefaultNetworkId
		bzzconfig.DeliverySkipCheck = true
		bzzconfig.SyncingSkipCheck = true
		bzzconfig.Path = bzzdir
		bzzconfig.Pss.AllowRaw = true
		bzzconfig.Init(prvkey)
		bzzconfig.Port = fmt.Sprintf("%d", bzzport)

		svc, err := swarm.NewSwarm(bzzconfig, nil)
		if err != nil {
			Log.Crit("unable to configure swarm", "err", err)
			return nil, err
		}

		for i, s := range specs {
			p, err := svc.RegisterPssProtocol(s, protocols[i], &pss.ProtocolParams{
				Asymmetric: true,
				Symmetric:  true,
			})
			if err != nil {
				return nil, err
			}

			// register topic for this protocol
			if topic != nil {
				p.Pss.Register(topic, p.Handle)
			}
			// append to pssprotos
			if pssprotos != nil {
				*pssprotos = append(*pssprotos, p)
				// LogError("Add item", "len", len(*pssprotos))
			}

		}

		return svc, nil
	}
}

// NewSwarmServiceWithProtocol : create new swarm service with protocol and topic
func NewSwarmServiceWithProtocol(stack *node.Node, bzzport int, specs []*protocols.Spec, protocols []*p2p.Protocol, topic *pss.Topic, pssprotos *[]*pss.Protocol) func(ctx *node.ServiceContext) (node.Service, error) {
	// get the encrypted private key file
	// load the private key from the file content ?
	prvkey, _ := crypto.GenerateKey()
	return NewSwarmServiceWithProtocolAndPrivateKey(stack, bzzport, specs, protocols, topic, pssprotos, prvkey)
}

func LogInfo(msg string, ctx ...interface{}) {
	Log.Info(color.HiGreenString(msg), ctx...)
}

func LogCrit(msg string, ctx ...interface{}) {
	Log.Info(color.HiRedString(msg), ctx...)
}

func LogDebug(msg string, ctx ...interface{}) {
	Log.Info(color.HiMagentaString(msg), ctx...)
}

func LogTrace(msg string, ctx ...interface{}) {
	Log.Trace(color.HiCyanString(msg), ctx...)
}

func LogError(msg string, ctx ...interface{}) {
	Log.Info(color.HiRedString(msg), ctx...)
}

func LogWarn(msg string, ctx ...interface{}) {
	Log.Info(color.HiYellowString(msg), ctx...)
}

func ToJSON(obj interface{}) string {
	bytes, _ := json.Marshal(obj)
	return string(bytes)
}

func NewServiceNodeWithPrivateKey(privateKey *ecdsa.PrivateKey, port int, httpport int, wsport int, modules ...string) (*node.Node, error) {
	if port == 0 {
		port = P2pPort
	}
	cfg := &node.DefaultConfig
	cfg.P2P.ListenAddr = fmt.Sprintf(":%d", port)
	cfg.P2P.EnableMsgEvents = true
	cfg.P2P.NoDiscovery = true
	cfg.IPCPath = IPCName
	cfg.DataDir = fmt.Sprintf("%s%d", DatadirPrefix, port)
	if privateKey != nil {
		cfg.P2P.PrivateKey = privateKey
	}
	if httpport > 0 {
		cfg.HTTPHost = node.DefaultHTTPHost
		cfg.HTTPPort = httpport
		cfg.HTTPCors = []string{"*"}
		cfg.HTTPModules = append(cfg.HTTPModules, modules...)
	}
	if wsport > 0 {
		cfg.WSHost = node.DefaultWSHost
		cfg.WSPort = wsport
		cfg.WSOrigins = []string{"*"}
		cfg.WSModules = append(cfg.WSModules, modules...)
	}
	stack, err := node.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("ServiceNode create fail: %v", err)
	}
	return stack, nil
}

// NewServiceNode : set up the local service node
func NewServiceNode(port int, httpport int, wsport int, modules ...string) (*node.Node, error) {
	return NewServiceNodeWithPrivateKey(nil, port, httpport, wsport, modules...)
}

// NewServer : create a server
func NewServer(privkey *ecdsa.PrivateKey, name string, version string, proto *p2p.Protocol, port int) *p2p.Server {

	cfg := p2p.Config{
		PrivateKey:      privkey,
		Name:            common.MakeName(name, version),
		MaxPeers:        1,
		EnableMsgEvents: true,
	}

	if proto != nil {
		cfg.Protocols = []p2p.Protocol{*proto}
	}

	if port > 0 {
		cfg.ListenAddr = fmt.Sprintf(":%d", port)
	}

	srv := &p2p.Server{
		Config: cfg,
	}
	return srv
}

// utility functions
//
// connects to the RPC specified by the url string
// on successful connection it retrieves the enode string from the endpoint
// RPC url can be IPC (filepath), websockets (ws://) or HTTP (http://)
func getEnodeFromRPC(rawurl string) (string, error) {
	rpcclient, err := rpc.Dial(rawurl)
	if err != nil {
		return "", fmt.Errorf("cannot add remote host: %v", err)
	}

	var nodeinfo p2p.NodeInfo
	err = rpcclient.Call(&nodeinfo, "admin_nodeInfo")
	if err != nil {
		return "", fmt.Errorf("RPC nodeinfo call failed: %v", err)
	}
	return nodeinfo.Enode, nil
}

// WaitHealthy : wait for nodes
func WaitHealthy(ctx context.Context, minbinsize int, rpcs ...*rpc.Client) error {
	var ids []discv5.NodeID
	var addrs [][]byte
	for _, r := range rpcs {
		var nodeinfo p2p.NodeInfo
		err := r.Call(&nodeinfo, "admin_nodeInfo")
		if err != nil {
			return err
		}
		p2pnode, err := discv5.ParseNode(nodeinfo.Enode)
		if err != nil {
			return err
		}
		ids = append(ids, p2pnode.ID)
		var bzzaddr string
		err = r.Call(&bzzaddr, "pss_baseAddr")
		if err != nil {
			return err
		}
		addrs = append(addrs, common.FromHex(bzzaddr))
	}
	peerpot := network.NewPeerPotMap(minbinsize, addrs)
	for {
		healthycount := 0
		for i, r := range rpcs {
			var health network.Health
			err := r.Call(&health, "hive_healthy", peerpot)
			if err != nil {
				return err
			}
			Log.Debug("health", "i", i, "addr", common.ToHex(addrs[i]), "id", ids[i], "info", health)
			if health.KnowNN && health.GotNN && health.Full {
				healthycount++
			}
		}
		if healthycount == len(rpcs) {
			break
		}
		time.Sleep(time.Millisecond * 250)
	}
	return nil
}

func WaitForCtrlC() {
	var endWaiter sync.WaitGroup
	endWaiter.Add(1)
	var signalChannel chan os.Signal
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		<-signalChannel
		endWaiter.Done()
	}()
	endWaiter.Wait()
}

func RandomID() (id enode.ID) {
	for i := range id {
		id[i] = byte(rand.Intn(255))
	}
	return id
}
