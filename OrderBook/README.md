## How to start

Copy or link this folder to GOPATH/src  
`ln -SF $PWD $GOPATH/src/github.com/tomochain/orderbook`

By default we use POA consensus for demo  
Assume you are in root folder  
Node1: `yarn node1 -mining true`  
Node2: `yarn node2 -mining true`  
Backend: `yarn backend`
