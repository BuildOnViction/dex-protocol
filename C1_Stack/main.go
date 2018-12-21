// set up boilerplate service node and start it
package main

import (
	"os"

	demo "github.com/tomochain/orderbook/common"
)

func main() {
	// set up the service node
	// create the node instance with the config
	stack, err := demo.NewServiceNode(0, 0, 0)
	if err != nil {
		demo.LogCrit("ServiceNode create fail", "err", err)
	}

	// start the node
	err = stack.Start()
	if err != nil {
		demo.LogCrit("ServiceNode start fail", "err", err)
	}
	defer os.RemoveAll(stack.DataDir())

	// shut down
	err = stack.Stop()
	if err != nil {
		demo.LogCrit("Node stop fail", "err", err)
	}
}
