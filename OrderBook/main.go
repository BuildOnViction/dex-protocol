// pss RPC routed over swarm
package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/downloader"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/discover"

	"github.com/manifoldco/promptui"
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

	// pssprotos = []*pss.Protocol{}
	// get the incoming message
	msgC = make(chan interface{})

	prompt          *promptui.Select
	commands        []terminal.Command
	genesisPath     string
	mining          bool
	orderbookEngine *orderbook.Engine
	nodeaddr        string
)

func initPrompt(privateKeyName string) {

	// default value for node2 if using keystore1 and vice versa
	if privateKeyName == "keystore1" {
		nodeaddr = "enode://ce24c4f944a0a3614b691d839a6a89339d17abac3d69c0d24e806db45d1bdbe7afa53c02136e5ad952f43e6e7285cd3971e367d8789f4eb7306770f5af78755d@127.0.0.1:30101?discport=0"
	} else {
		nodeaddr = "enode://655b231711df566a1bbf8f62dd0abaad71a1baa2c4bc865cae1691431bff2d9185fb66c99b982e20fd0fd562ced2c1ced96bd3e1daba0235870dfce0663a3483@127.0.0.1:30100?discport=0"
	}

	orderArguments := []terminal.Argument{
		{Name: "id", Value: "1"},
		{Name: "pairName", Value: "TOMO/WETH"},
		{Name: "type", Value: "limit"},
		{Name: "side", Value: orderbook.Ask},
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
		// {
		// 	Name:        "publicKey",
		// 	Description: "Get public key",
		// },
		// {
		// 	Name: "addNode",
		// 	Arguments: []terminal.Argument{
		// 		{Name: "nodeaddr", Value: nodeaddr},
		// 	},
		// 	Description: "Add node to seed",
		// },
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
				mining = c.Bool("mining")
				// init prompt
				initPrompt(privateKeyName)
				// must return export function
				return Start(c.Int("p2pPort"), c.Int("httpPort"), c.Int("wsPort"), c.String("name"),
					c.String("privateKey"))
			},
			Flags: []cli.Flag{
				cli.IntFlag{Name: "p2pPort, p1", Value: demo.P2pPort},
				cli.IntFlag{Name: "httpPort, p2", Value: node.DefaultHTTPPort},
				cli.IntFlag{Name: "wsPort, p3", Value: demo.WSDefaultPort},
				cli.StringFlag{Name: "name, n", Value: "node1"},
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

func Start(p2pPort int, httpPort int, wsPort int, name string, privateKey string) error {

	// start the program at other rtine
	_, fileName, _, _ := runtime.Caller(1)
	basePath := path.Dir(fileName)
	privateKeyPath := path.Join(basePath, privateKey)
	genesisPath = path.Join(basePath, "genesis.json")
	// privateKeyPath is from current folder where the file is running
	demo.LogInfo("Start node", "host", node.DefaultWSHost,
		"p2pPort", p2pPort, "httpPort", httpPort, "wsPort", wsPort, "name",
		name, "privateKey", privateKeyPath)

	startup(p2pPort, httpPort, wsPort, name, privateKeyPath)

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
				// commit changes to orderbook
				orderbookEngine.Commit()
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

			// case "addNode":
			// 	nodeaddr := results["nodeaddr"]
			// 	demo.LogInfo(fmt.Sprintf("-> Add node: %s\n", nodeaddr))
			// 	addNode(nodeaddr)

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

func addNode(rawurl string) error {

	newNode, err := discover.ParseNode(rawurl)
	if err != nil {
		demo.LogCrit("pass node addr fail", "err", err)
		return err
	}

	demo.LogInfo("add node", "node", newNode.String())
	thisNode.Server().AddPeer(newNode)

	demo.LogInfo("Added node successfully!")
	return nil
}

func nodeAddr() string {
	return thisNode.Server().Self().String()
}

func processOrder(payload map[string]string) error {
	// add order at this current node first
	// get timestamp in milliseconds
	payload["timestamp"] = strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	msg, err := protocol.NewOrderbookMsg(payload)
	if err == nil {
		// try to store into model, if success then process at local and broad cast
		trades, orderInBook := orderbookEngine.ProcessOrder(payload)
		demo.LogInfo("Orderbook result", "Trade", trades, "OrderInBook", orderInBook)

		// broad cast message
		msgC <- msg

	}

	return nil
}

func updateEthService() {
	// make full node after start the node, then later register swarm service over that node
	ethConfig, err := initGenesis(thisNode)
	ethConfig.Etherbase = crypto.PubkeyToAddress(privkey.PublicKey)
	if err != nil {
		panic(err.Error())
	}

	// register ethservice with genesis block
	utils.RegisterEthService(thisNode, ethConfig)

	if err != nil {
		demo.LogCrit("servicenode pss register fail", "err", err)
	}

	// config ethereum
	var ethereum *eth.Ethereum
	if err := thisNode.Service(&ethereum); err != nil {
		demo.LogError(fmt.Sprintf("Ethereum service not running: %v", err))
	}

	// config ethereum gas price
	ethereum.TxPool().SetGasPrice(ethConfig.GasPrice)

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
		miner := ethereum.Miner()
		miner.Start(account.Address)
	}
}

func updateExtraService(rpcapi []string) {

	for _, api := range rpcapi {
		if api == "eth" {
			updateEthService()
		}
	}

}

// simple ping and receive protocol
func startup(p2pPort int, httpPort int, wsPort int, name string, privateKey string) {

	var err error

	// get private key
	privkey, err = crypto.LoadECDSA(privateKey)

	// register pss and orderbook service
	rpcapi := []string{
		// "eth",
		// "ssh",
		"orderbook",
	}
	dataDir := fmt.Sprintf("%s%d", demo.DatadirPrefix, p2pPort)
	orderbookDir := path.Join(dataDir, "orderbook")
	allowedPairs := map[string]*big.Int{
		"TOMO/WETH": big.NewInt(10e9),
	}
	orderbookEngine = orderbook.NewEngine(orderbookDir, allowedPairs)

	thisNode, err = demo.NewServiceNodeWithPrivateKeyAndDataDir(privkey, dataDir, p2pPort, httpPort, wsPort, rpcapi...)
	// register normal service, protocol is for p2p, service is for rpc calls
	service := protocol.NewService(msgC, quitC, orderbookEngine)
	err = thisNode.Register(service)

	if err != nil {
		demo.LogCrit("Register orderbook service in servicenode failed", "err", err)
	}

	if err != nil {
		demo.LogCrit(err.Error())
	}

	// extra service like eth, swarm need more work
	updateExtraService(rpcapi)

	// start the nodes
	err = thisNode.Start()
	if err != nil {
		demo.LogCrit("servicenode start failed", "err", err)
	}
	addNode(nodeaddr)
}

// geth init genesis.json --datadir .datadir
func initGenesis(stack *node.Node) (*eth.Config, error) {
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

	demo.LogInfo("Genesis", "chainid", genesis.Config.ChainId.String())
	return ethConfig, err
}
