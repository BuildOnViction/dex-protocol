// pss RPC routed over swarm
package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/swarm/storage/feed"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/node"
	"github.com/manifoldco/promptui"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/swarm/pss"
	demo "github.com/tomochain/orderbook/common"
	"github.com/tomochain/orderbook/orderbook"
	"github.com/tomochain/orderbook/protocol"
	"github.com/tomochain/orderbook/terminal"
	cli "gopkg.in/urfave/cli.v1"
)

// command line arguments
var (
	thisNode *node.Node
	// struct{} is the smallest data type available in Go, since it contains literally nothing
	quitC   = make(chan struct{}, 1)
	app     = cli.NewApp()
	privkey *ecdsa.PrivateKey

	pssprotos = []*pss.Protocol{}
	// get the incoming message
	msgC = make(chan interface{})

	prompt   *promptui.Select
	commands []terminal.Command

	orderbookModel *protocol.OrderbookModel
)

func initPrompt(privateKeyName string) {

	// default value for node2 if using keystore1 and vice versa
	var bzzaddr, nodeaddr, publickey string
	if privateKeyName == "keystore1" {
		bzzaddr = "0x9984c9556ca87842c4ceb839518cd3648dc495d579f7af7f9ba49989bc207346"
		nodeaddr = "enode://ce24c4f944a0a3614b691d839a6a89339d17abac3d69c0d24e806db45d1bdbe7afa53c02136e5ad952f43e6e7285cd3971e367d8789f4eb7306770f5af78755d@127.0.0.1:30101?discport=0"
		publickey = "0x04ce24c4f944a0a3614b691d839a6a89339d17abac3d69c0d24e806db45d1bdbe7afa53c02136e5ad952f43e6e7285cd3971e367d8789f4eb7306770f5af78755d"
	} else {
		bzzaddr = "0x1bb065aa5e7997efc322a0223b62cd4ca218ce9f51c1e85bf9bd429f9265a3d7"
		nodeaddr = "enode://655b231711df566a1bbf8f62dd0abaad71a1baa2c4bc865cae1691431bff2d9185fb66c99b982e20fd0fd562ced2c1ced96bd3e1daba0235870dfce0663a3483@127.0.0.1:30100?discport=0"
		publickey = "0x04655b231711df566a1bbf8f62dd0abaad71a1baa2c4bc865cae1691431bff2d9185fb66c99b982e20fd0fd562ced2c1ced96bd3e1daba0235870dfce0663a3483"
	}

	orderArguments := []terminal.Argument{
		{Name: "id", Value: "1"},
		{Name: "coin", Value: "Tomo"},
		{Name: "type", Value: "limit"},
		{Name: "side", Value: orderbook.ASK},
		{Name: "quantity", Value: "10"},
		{Name: "price", Value: "100", Hide: func(results map[string]string, thisArgument *terminal.Argument) bool {
			// ignore this argument when order type is market
			if results["type"] == "market" {
				return true
			}
			return false
		}},
		{Name: "trade_id", Value: "1"},
	}

	// init prompt commands
	commands = []terminal.Command{
		{
			Name:        "processOrder",
			Arguments:   orderArguments,
			Description: "Process order and store on swarm network",
		},
		{
			Name:        "publicKey",
			Description: "Get public key",
		},
		{
			Name: "addNode",
			Arguments: []terminal.Argument{
				{Name: "nodeaddr", Value: nodeaddr},
				{Name: "publickey", Value: publickey},
			},
			Description: "Add node to seed",
		},
		{
			Name: "updateBzz",
			Arguments: []terminal.Argument{
				{Name: "bzzaddr", Value: bzzaddr},
				{Name: "publickey", Value: publickey},
			},
			Description: "Associate swarm topic with public key",
		},
		{
			Name:        "bzzAddr",
			Description: "Get Swarm address",
		},
		{
			Name:        "nodeAddr",
			Description: "Get Node address",
		},
		{
			Name:        "quit",
			Description: "Quit the program",
		},
	}

	// cast type to sort
	// sort.Sort(terminal.CommandsByName(commands))

	prompt = terminal.NewPrompt("Your choice:", 6, commands)
}

func init() {
	// Initialize the CLI app and start tomo
	app.Commands = []cli.Command{
		cli.Command{
			Name: "start",
			Action: func(c *cli.Context) error {
				privateKeyName := path.Base(c.String("privateKey"))
				// init prompt
				initPrompt(privateKeyName)
				// must return export function
				return Start(c.Int("p2pPort"), c.Int("httpPort"), c.Int("wsPort"), c.Int("bzzPort"), c.String("privateKey"), c.Bool("mining"))
			},
			Flags: []cli.Flag{
				cli.IntFlag{Name: "p2pPort, p1", Value: demo.P2pPort},
				cli.IntFlag{Name: "httpPort, p2", Value: node.DefaultHTTPPort},
				cli.IntFlag{Name: "wsPort, p3", Value: demo.WSDefaultPort},
				cli.IntFlag{Name: "bzzPort, p4", Value: demo.BzzDefaultPort},
				cli.StringFlag{Name: "privateKey, pvk"},
				cli.BoolFlag{Name: "mining, m"},
			},
		},
	}

}

func main() {

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func Start(p2pPort int, httpPort int, wsPort int, bzzPort int, privateKey string, mining bool) error {

	// start the program at other rtine
	_, fileName, _, _ := runtime.Caller(1)
	basePath := path.Dir(fileName)
	privateKeyPath := path.Join(basePath, privateKey)
	genesisPath := path.Join(basePath, "genesis.json")
	// privateKeyPath is from current folder where the file is running
	demo.LogInfo("Connecting to pss websocket", "host", node.DefaultWSHost,
		"p2pPort", p2pPort, "httpPort", httpPort, "wsPort", wsPort, "bzzPort",
		bzzPort, "privateKey", privateKeyPath)

	startup(p2pPort, httpPort, wsPort, bzzPort, privateKeyPath, genesisPath, mining)

	// process command
	fmt.Println("---------------Welcome to Orderbook over swarm testing---------------------")
	var endWaiter sync.WaitGroup
	endWaiter.Add(1)

	// start serving
	go func() {

		for {
			// loop command
			selected, _, err := prompt.Run()

			// unknow error, should retry
			if err != nil {
				demo.LogInfo("Prompt failed %v\n", err)
				continue
			}

			// get selected command and run it
			command := commands[selected]
			if command.Name == "quit" {
				demo.LogInfo("Server quiting...")
				endWaiter.Done()
				thisNode.Stop()
				quitC <- struct{}{}
				demo.LogInfo("-> Goodbye\n")
				return
			}
			results := command.Run()

			// process command
			switch command.Name {
			case "processOrder":
				demo.LogInfo("-> Add order", "payload", results)
				// put message on channel
				go processOrder(results)
			case "publicKey":
				demo.LogInfo(fmt.Sprintf("-> Public Key: %s\n", publicKey()))
			case "updateBzz":
				bzzaddr := results["bzzaddr"]
				publickey := results["publickey"]
				demo.LogInfo(fmt.Sprintf("-> Add bzz: %s\n", bzzaddr))
				updateBzz(bzzaddr, publickey)
			case "addNode":
				nodeaddr := results["nodeaddr"]
				publickey := results["publickey"]
				demo.LogInfo(fmt.Sprintf("-> Add node: %s\n", nodeaddr))
				addNode(nodeaddr, publickey)
			case "bzzAddr":
				demo.LogInfo(fmt.Sprintf("-> BZZ Address: %s\n", bzzAddr()))
			case "nodeAddr":
				demo.LogInfo(fmt.Sprintf("-> Node Address: %s\n", nodeAddr()))

			default:
				demo.LogInfo(fmt.Sprintf("-> Unknown command: %s\n", command.Name))
			}
		}

	}()

	// wait for command processing
	endWaiter.Wait()

	// finally shutdown
	return shutdown()
}

func shutdown() error {
	// return os.RemoveAll(thisNode.DataDir())
	return nil
}

func updateBzz(bzzAddr string, publicKey string) error {
	rpcClient, err := thisNode.Attach()

	// Set Public key to associate with a particular Pss peer
	err = rpcClient.Call(nil, "pss_setPeerPublicKey", publicKey, protocol.OrderbookTopic, bzzAddr)
	return err
}

func addNode(rawurl string, publicKey string) error {

	newNode, err := enode.ParseV4(rawurl)
	if err != nil {
		demo.LogCrit("pass node addr fail", "err", err)
		return err
	}

	demo.LogInfo("add node", "node", newNode.String())
	thisNode.Server().AddPeer(newNode)

	// if have protocol implemented
	if len(pssprotos) > 0 {
		nid := newNode.ID()
		p := p2p.NewPeer(nid, nid.String(), []p2p.Cap{})

		// add peer with its public key to this protocol on a topic and using asymetric cryptography
		pssprotos[0].AddPeer(p, protocol.OrderbookTopic, true, publicKey)

		if err != nil {
			demo.LogCrit("pss set pubkey fail", "err", err)
		}

	}

	demo.LogInfo("Added node successfully!")
	return nil
}

func nodeAddr() string {
	return thisNode.Server().Self().String()
}

func bzzAddr() string {
	// get the rpc clients
	rpcClient, err := thisNode.Attach()
	// get the recipient node's swarm overlay address
	var bzzAddr string
	err = rpcClient.Call(&bzzAddr, "pss_baseAddr")
	if err != nil {
		demo.LogCrit("pss get baseaddr fail", "err", err)
	}
	return bzzAddr
}

func publicKey() string {
	// get the publickeys
	var pubkey string
	rpcClient, err := thisNode.Attach()
	err = rpcClient.Call(&pubkey, "pss_getPublicKey")
	if err != nil {
		demo.LogCrit("pss get pubkey fail", "err", err)
	}
	return pubkey
}

func processOrder(payload map[string]string) error {
	// add order at this current node first
	// get timestamp in milliseconds
	payload["timestamp"] = strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	msg, err := protocol.NewOrderbookMsg(payload)
	if err == nil {
		// try to store into model, if success then process at local and broad cast
		err = orderbookModel.ProcessOrder(msg)
		if err == nil {
			trades, orderInBook := orderbookModel.Orderbook.ProcessOrder(payload, false)
			demo.LogInfo("Orderbook result", "Trade", trades, "OrderInBook", orderInBook)
			msgC <- msg
		}
	}

	return nil
}

// simple ping and receive protocol
func startup(p2pPort int, httpPort int, wsPort int, bzzPort int, privateKey string, genesisPath string, mining bool) {

	var err error

	// get private key
	privkey, err = crypto.LoadECDSA(privateKey)

	// register pss and orderbook service
	rpcapi := []string{
		"eth",
		// "ssh",
		"personal",
		"pss",
		"orderbook",
	}
	thisNode, err = demo.NewServiceNodeWithPrivateKey(privkey, p2pPort, httpPort, wsPort, rpcapi...)

	if err != nil {
		demo.LogCrit(err.Error())
	}

	// make full node after start the node, then later register swarm service over that node
	ethConfig, err := initGenesis(thisNode, genesisPath)
	ethConfig.Etherbase = crypto.PubkeyToAddress(privkey.PublicKey)
	if err != nil {
		panic(err.Error())
	}

	// register ethservice with genesis block
	utils.RegisterEthService(thisNode, ethConfig)

	// do we need to register whisper? maybe not
	// sshConfig := whisperv6.DefaultConfig
	// utils.RegisterShhService(thisNode, &sshConfig)

	// register swarm service
	bzzURL := fmt.Sprintf("http://%s:%d", node.DefaultHTTPHost, bzzPort)

	signer := feed.NewGenericSigner(privkey)
	orderbookModel = protocol.NewModel(bzzURL, signer)

	protocolSpecs := []*protocols.Spec{protocol.OrderbookProtocol}
	proto := protocol.New(msgC, quitC, orderbookModel)
	protocolArr := []*p2p.Protocol{proto}

	// register the pss activated bzz services
	svc := demo.NewSwarmServiceWithProtocolAndPrivateKey(thisNode, bzzPort, protocolSpecs, protocolArr, &protocol.OrderbookTopic, &pssprotos, privkey)
	err = thisNode.Register(svc)

	if err != nil {
		demo.LogCrit("servicenode pss register fail", "err", err)
	}
	// register normal service, using bzz client internally
	err = thisNode.Register(protocol.NewService(orderbookModel))
	if err != nil {
		demo.LogCrit("Register orderbook service in servicenode failed", "err", err)
	}

	// start the nodes
	err = thisNode.Start()
	if err != nil {
		demo.LogCrit("servicenode start failed", "err", err)
	}

	rpcClient, err := thisNode.Attach()
	// wait until the state of the swarm overlay network is ready
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = demo.WaitHealthy(ctx, 1, rpcClient)
	if err != nil {
		demo.LogCrit("health check fail", "err", err)
	}
	time.Sleep(time.Second) // because the healthy does not work

	// config ethereum
	var ethereum *eth.Ethereum
	if err := thisNode.Service(&ethereum); err != nil {
		demo.LogError(fmt.Sprintf("Ethereum service not running: %v", err))
	}

	// config ethereum gas price
	ethereum.TxPool().SetGasPrice(ethConfig.MinerGasPrice)

	password := "123456789"
	ks := thisNode.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	// import will create a keystore if not found
	account, err := ks.ImportECDSA(privkey, password)
	err = ks.Unlock(account, password)
	if err != nil {
		panic(err)
	}

	// if this node can mine
	if mining {
		if err := ethereum.StartMining(0); err != nil {
			demo.LogError(fmt.Sprintf("Failed to start mining: %v", err))
		}
	}

}

// geth init genesis.json --datadir .datadir
func initGenesis(stack *node.Node, genesisPath string) (*eth.Config, error) {
	ethConfig := &eth.DefaultConfig
	ethConfig.SyncMode = downloader.FullSync
	ethConfig.SkipBcVersionCheck = true
	ethConfig.NetworkId = 8888

	var genesis core.Genesis
	databytes, err := ioutil.ReadFile(genesisPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(databytes, &genesis)
	ethConfig.Genesis = &genesis

	demo.LogInfo("Genesis", "chainid", genesis.Config.ChainID)
	return ethConfig, err
}
