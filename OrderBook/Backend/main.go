// pss RPC routed over swarm
package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	demo "../../common"
	"../protocol"
	"../terminal"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/manifoldco/promptui"
	"gopkg.in/urfave/cli.v1"
)

var (
	app       = cli.NewApp()
	rpcClient *rpc.Client
	prompt    *promptui.Select
	commands  []terminal.Command
)

func init() {
	// Initialize the CLI app and start tomo
	app.Commands = []cli.Command{
		cli.Command{
			Name: "rpc",
			Action: func(c *cli.Context) error {
				// must return export function
				return Start()
			},
			Flags: []cli.Flag{
				cli.IntFlag{Name: "wsPort, p", Value: demo.WSDefaultPort},
			},
		},
		cli.Command{
			Name: "savekey",
			Action: func(c *cli.Context) error {
				// must return export function
				return SaveKey(c.String("path"))
			},
			Flags: []cli.Flag{
				cli.StringFlag{Name: "path, p", Value: "../keystore"},
			},
		},
		cli.Command{
			Name: "loadkey",
			Action: func(c *cli.Context) error {
				// must return export function
				return LoadKey(c.String("path"))
			},
			Flags: []cli.Flag{
				cli.StringFlag{Name: "path, p", Value: "../keystore"},
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

func initPrompt() {
	// init prompt commands
	commands = []terminal.Command{
		{
			Name: "getOrders",
			Arguments: []terminal.Argument{
				{Name: "coin", Value: "Tomo"},
				{Name: "signerAddress", Value: "0x28074f8d0fd78629cd59290cac185611a8d60109"},
			},
			Description: "Get the order from the swarm storgae",
		},
		{
			Name: "updatePort",
			Arguments: []terminal.Argument{
				{Name: "wsPort", Value: "18543", AllowEdit: true},
			},
			Description: "Update the websocket port to call RPC",
		},
		// {
		// 	Name: "updateOrder",
		// 	Arguments: []terminal.Argument{
		// 		{Name: "orderID", Value: "1"},
		// 		{Name: "type", Value: "limit"},
		// 		{Name: "side", Value: "ask"},
		// 		{Name: "quantity", Value: "10"},
		// 		{Name: "price", Value: "100", Hide: func(results map[string]string) bool {
		// 			// ignore this argument when order type is market
		// 			if results["type"] == "market" {
		// 				return true
		// 			}
		// 			return false
		// 		}},
		// 		{Name: "trade_id", Value: "1"},
		// 	},
		// 	Description: "Get the order from the swarm storgae",
		// },
		{
			Name:        "getBestAskList",
			Description: "Get best ask list",
		},
		{
			Name:        "getBestBidList",
			Description: "Get best bid list",
		},
		{
			Name:        "quit",
			Description: "Quit the program",
		},
	}
	// sort.Sort(terminal.CommandsByName(commands))
	prompt = terminal.NewPrompt("Your choice:", 4, commands)
}

func SaveKey(path string) error {

	privkey, _ := crypto.GenerateKey()
	return crypto.SaveECDSA(path, privkey)

}

func LoadKey(path string) error {
	privkey, _ := crypto.LoadECDSA(path)
	demo.LogInfo("privkey", "publickey", privkey.PublicKey)
	return nil
}

func logResult(result interface{}, err error) {
	if err != nil {
		demo.LogCrit("RPC call fail", "err", err)
	} else {
		demo.LogInfo("Get response", "result", result)
	}
}
func callRPC(result interface{}, function string, params ...interface{}) {
	// assume there is no argument at all
	err := rpcClient.Call(&result, function, params...)
	demo.LogInfo("Call", "function", function, "params", params)
	logResult(result, err)
}

func Start() error {

	initPrompt()

	// configure and start up pss client RPCs
	// we can use websockets ...

	// get a valid topic byte
	// get a valid topic byte
	// call the RPC method, will be name_methodInCammelCase
	// process command
	fmt.Println("---------------Welcome to Backend testing---------------------")
	var endWaiter sync.WaitGroup
	endWaiter.Add(1)

	// start serving
	go func() {
		var wsPort = "18543"
		var signerAddress = "0x28074f8d0fd78629cd59290cac185611a8d60109"
		for {
			// loop command
			commands[1].Arguments[0].Value = wsPort
			if wsPort != "18543" {
				signerAddress = "0x6e6BB166F420DDd682cAEbf55dAfBaFda74f2c9c"
			}
			commands[0].Arguments[1].Value = signerAddress

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
				demo.LogInfo("-> Goodbye\n")
				return
			}
			results := command.Run()

			// wait until the state of the swarm overlay network is ready
			endpoint := fmt.Sprintf("ws://%s:%s", node.DefaultWSHost, wsPort)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			rpcClient, err = rpc.DialWebsocket(ctx, endpoint, "*")
			// rightRPCClient, err := rpc.DialWebsocket(ctx, rightEndpoint, "*")
			if err != nil {
				demo.LogCrit("WS RPC dial failed", "err", err)
				cancel()
				continue
			}
			// process command
			var result interface{}
			switch command.Name {
			case "updatePort":
				demo.LogInfo("-> Update", "wsPort", results["wsPort"])
				wsPort = results["wsPort"]
			case "getOrders":
				demo.LogInfo("-> Get orders", "coin", results["coin"], "from", results["signerAddress"])
				// put message on channel
				var orderResult []protocol.OrderbookMsg
				err := rpcClient.Call(&orderResult, "orderbook_getOrders", results["coin"], results["signerAddress"])
				logResult(orderResult, err)
			case "getBestAskList":
				demo.LogInfo("-> Best ask list:")
				callRPC(result, "orderbook_getBestAskList")
			case "getBestBidList":
				demo.LogInfo("-> Best bid list:")
				callRPC(result, "orderbook_getBestBidList")
			default:
				demo.LogInfo(fmt.Sprintf("-> Unknown command: %s\n", command.Name))
			}

		}
	}()

	// wait for command processing
	endWaiter.Wait()

	// bring down the servicenodes
	if rpcClient != nil {
		rpcClient.Close()
	}

	return nil
}
