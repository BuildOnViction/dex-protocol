## Update Orderlist flow

![diagram1](./images/diagram1.svg)

## Overall data flow

![diagram2](./images/diagram2.svg)

## Swarm usage

1. Only Feed storage and message queue is enable (no file upload to spam)
1. Only User has coin can order, and update his own orders
1. A user can only attemp a fix number of orders, and update an order in a specific amount of time base on timestamp of each order
1. Each topic is a coin symbol, so the total of topics is limited
1. First time new node needs to crawl orders to catchup, then subscribes to the order channel for realtime update

## How to start

Copy or link this folder to GOPATH/src  
`ln -SF $PWD $GOPATH/src/github.com/tomochain/orderbook`

By default we use POA consensus for demo  
Assume you are in root folder  
Node1: `yarn node1 -mining true`  
Node2: `yarn node2 -mining true`  
Backend: `yarn backend`

## TODO

1. Move swarm to a new package named ethereum-swarm and use tomochain as the original go-ethereum.
2. Optimize swarm to make it fit better with the running version
3. Wrap all orderbook APIs using rpcClient into orderbookClient
