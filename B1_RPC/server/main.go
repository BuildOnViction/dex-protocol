// RPC hello world
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"

	demo "github.com/tomochain/orderbook/common"
)

// set up an object that can contain the API methods
// these API methods can be used for orderbook
type CrudAPI struct {
	V int
}

var database = make(map[string]string)

// a valid API method is exported, has a pointer receiver and returns error as last argument
// the method will be called with <registeredname>_helloWorld
// (first letter in method is lowercase, module name and method name separated by underscore)
func (api *CrudAPI) Put(key, value string) (string, error) {
	demo.LogInfo("Get put request", "key", key, "value", value)
	database[key] = value
	return fmt.Sprintf("Put %s success", key), nil
}

func (api *CrudAPI) Get(keys []string) ([]string, error) {
	demo.LogInfo("Get read request", "keys", keys)
	var result []string
	for _, key := range keys {
		result = append(result, database[key])
	}
	return result, nil
}

func (api *CrudAPI) Delete(key string) (string, error) {
	demo.LogInfo("Get delete request", "key", key)
	delete(database, key)
	return fmt.Sprintf("Delete %s success", key), nil
}

func main() {

	// set up the RPC server
	timeouts := rpc.DefaultHTTPTimeouts
	vhost := node.DefaultConfig.HTTPVirtualHosts
	cors := []string{"*"}
	rpcAPIs := []rpc.API{
		{
			Namespace: "crud",
			Public:    true,
			Service: &CrudAPI{
				V: 42,
			},
			Version: "1.0",
		},
	}
	modules := []string{"crud"}
	address := fmt.Sprintf("localhost:%d", node.DefaultHTTPPort)
	listener, rpcsrv, err := rpc.StartHTTPEndpoint(address, rpcAPIs, modules, cors, vhost, timeouts)
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
