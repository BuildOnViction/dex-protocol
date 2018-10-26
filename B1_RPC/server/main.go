// RPC hello world
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"

	demo "../../common"
)

// set up an object that can contain the API methods
// these API methods can be used for orderbook
type FooAPI struct {
	V int
}

// a valid API method is exported, has a pointer receiver and returns error as last argument
// the method will be called with <registeredname>_helloWorld
// (first letter in method is lowercase, module name and method name separated by underscore)
func (api *FooAPI) HelloWorld(name string) (string, error) {
	demo.LogInfo("Get request", "name", name)
	return fmt.Sprintf("Hello %s", name), nil
}

func main() {

	// set up the RPC server
	timeouts := rpc.DefaultHTTPTimeouts
	vhost := node.DefaultConfig.HTTPVirtualHosts
	cors := []string{""}
	rpcAPIs := []rpc.API{
		{
			Namespace: "foo",
			Public:    true,
			Service: &FooAPI{
				V: 42,
			},
			Version: "1.0",
		},
	}
	modules := []string{"foo"}
	listener, rpcsrv, err := rpc.StartHTTPEndpoint("localhost:8454", rpcAPIs, modules, cors, vhost, timeouts)
	if err != nil {
		demo.LogCrit("Register API method(s) fail", "err", err)
	} else {
		demo.LogInfo("HTTP server is starting at", "address", listener.Addr())
	}

	fmt.Printf("Press Ctrl+C to end\n")
	demo.WaitForCtrlC()
	// bring down the server
	rpcsrv.Stop()
}
