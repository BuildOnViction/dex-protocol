// RPC hello world
package main

import (
	"fmt"
	"net"
	"os"

	"github.com/ethereum/go-ethereum/rpc"

	demo "../common"
)

// set up an object that can contain the API methods
// these API methods can be used for orderbook
type FooAPI struct {
}

// a valid API method is exported, has a pointer receiver and returns error as last argument
// the method will be called with <registeredname>_helloWorld
// (first letter in method is lowercase, module name and method name separated by underscore)
func (api *FooAPI) HelloWorld(name string) (string, error) {
	return fmt.Sprintf("Hello %s", name), nil
}

func main() {

	// set up the RPC server
	rpcsrv := rpc.NewServer()
	err := rpcsrv.RegisterName("foo", &FooAPI{})
	if err != nil {
		demo.LogCrit("Register API method(s) fail", "err", err)
	}

	// create IPC endpoint : interprocess communication work on your local computers
	ipcpath := ".demo.ipc"
	// using file socket with unix
	ipclistener, err := net.Listen("unix", ipcpath)
	if err != nil {
		demo.LogCrit("IPC endpoint create fail", "err", err)
	}
	defer os.Remove(ipcpath)

	// mount RPC server on IPC endpoint
	// it will automatically detect and serve any valid methods
	go func() {
		err = rpcsrv.ServeListener(ipclistener)
		if err != nil {
			demo.LogCrit("Mount RPC on IPC fail", "err", err)
		}
	}()

	// create an IPC client
	rpcclient, err := rpc.Dial(ipcpath)
	if err != nil {
		demo.LogCrit("IPC dial fail", "err", err)
	}

	// call the RPC method, will be name_methodInCammelCase
	var result string
	err = rpcclient.Call(&result, "foo_helloWorld", "Tu")
	if err != nil {
		demo.LogCrit("RPC call fail", "err", err)
	}

	// inspect the results
	demo.LogInfo("RPC return value", "reply", result)

	// bring down the server
	rpcsrv.Stop()
}
