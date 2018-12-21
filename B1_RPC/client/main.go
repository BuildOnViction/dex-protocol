// querying the p2p Server through RPC
package main

import (
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/manifoldco/promptui"
	demo "github.com/tomochain/orderbook/common"
	"github.com/tomochain/orderbook/terminal"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	app      = cli.NewApp()
	prompt   *promptui.Select
	commands []terminal.Command
)

func init() {
	// Initialize the CLI app and start tomo
	app.Commands = []cli.Command{
		cli.Command{
			Name: "start",
			Action: func(c *cli.Context) error {
				// must return export function
				return Start(c.Int("port"))
			},
			Flags: []cli.Flag{
				cli.IntFlag{Name: "port, p", Value: node.DefaultHTTPPort},
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
			Name: "put",
			Arguments: []terminal.Argument{
				{Name: "key", Value: "key"},
				{Name: "value", Value: "value"},
			},
			Description: "Put an item on network",
		},
		{
			Name: "get",
			Arguments: []terminal.Argument{
				{Name: "key", Value: "key", AllowEdit: true},
			},
			Description: "Get an item from network",
		},
		{
			Name: "delete",
			Arguments: []terminal.Argument{
				{Name: "key", Value: "key", AllowEdit: true},
			},
			Description: "Remove the item from network by its key",
		},
		{
			Name:        "quit",
			Description: "Quit the program",
		},
	}
	// sort.Sort(terminal.CommandsByName(commands))
	prompt = terminal.NewPrompt("Enter command:", 4, commands)
}

func Start(port int) error {
	rpcClient, err := rpc.DialHTTP(fmt.Sprintf("http://127.0.0.1:%d", port))
	if err != nil {
		demo.LogCrit("HTTP dial fail", "err", err)
		return nil
	}

	initPrompt()

	// configure and start up pss client RPCs
	// we can use websockets ...

	// get a valid topic byte
	// get a valid topic byte
	// call the RPC method, will be name_methodInCammelCase
	// process command
	fmt.Println("---------------Welcome to RPC testing---------------------")
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
				demo.LogInfo("Quiting...")
				endWaiter.Done()
				demo.LogInfo("-> Goodbye\n")
				return
			}
			results := command.Run()

			// process command
			var result interface{}
			switch command.Name {
			case "put":
				demo.LogInfo("-> Put item", "key", results["key"], "value", results["value"])
				var result interface{}
				err := rpcClient.Call(&result, "crud_put", results["key"], results["value"])
				demo.LogInfo(fmt.Sprintf("-> Result: %s, err: %s\n", result, err))
			case "get":
				demo.LogInfo("-> Get item", "key", results["key"])
				var result interface{}
				keys := regexp.MustCompile(`\s+`).Split(results["key"], -1)
				err := rpcClient.Call(&result, "crud_get", keys)
				demo.LogInfo(fmt.Sprintf("-> Result: %s, err: %s\n", result, err))
			case "delete":
				demo.LogInfo("-> Delete item", "key", results["key"])
				err := rpcClient.Call(&result, "crud_delete", results["key"])
				demo.LogInfo(fmt.Sprintf("-> Result: %s, err: %s\n", result, err))
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
