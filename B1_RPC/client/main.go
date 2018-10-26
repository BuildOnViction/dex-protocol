// querying the p2p Server through RPC
package main

import (
	"fmt"
	"os"

	demo "../../common"
	"github.com/ethereum/go-ethereum/rpc"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	app = cli.NewApp()
)

func init() {
	// Initialize the CLI app and start tomo
	app.Commands = []cli.Command{
		cli.Command{
			Name: "hello",
			Action: func(c *cli.Context) error {
				return Send(c.String("name"))
			},
			Flags: []cli.Flag{
				cli.StringFlag{Name: "name"},
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

// Send : send message to the server
func Send(str string) error {
	// create a TCP client
	rpcclient, err := rpc.DialHTTP("http://127.0.0.1:8454")
	if err != nil {
		demo.LogCrit("HTTP dial fail", "err", err)
		return nil
	}

	// call the RPC method, will be name_methodInCammelCase
	var result string

	err = rpcclient.Call(&result, "foo_helloWorld", str)
	if err != nil {
		demo.LogCrit("RPC call fail", "err", err)
	}

	// inspect the results
	demo.LogInfo("RPC return value", "reply", result)
	return nil
}
