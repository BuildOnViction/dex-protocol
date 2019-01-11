// pss RPC routed over swarm
package main

// command line arguments
var (
	bzzAddr string
)

// func main() {

// 	if err := app.Run(os.Args); err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 		os.Exit(1)
// 	}

// }

// func Start(p2pPort int, httpPort int, wsPort int, bzzPort int, privateKey string) error {

// 	// start the program at other rtine
// 	_, fileName, _, _ := runtime.Caller(1)
// 	basePath := path.Dir(fileName)
// 	privateKeyPath := path.Join(basePath, privateKey)
// 	genesisPath = path.Join(basePath, "genesis.json")
// 	// privateKeyPath is from current folder where the file is running
// 	demo.LogInfo("Start node", "host", node.DefaultWSHost,
// 		"p2pPort", p2pPort, "httpPort", httpPort, "wsPort", wsPort, "bzzPort",
// 		bzzPort, "privateKey", privateKeyPath)

// 	startup(p2pPort, httpPort, wsPort, bzzPort, privateKeyPath)

// 	// process command
// 	fmt.Println("---------------Welcome to Orderbook over swarm testing---------------------")
// 	var endWaiter sync.WaitGroup
// 	endWaiter.Add(1)

// 	// start serving
// 	go func() {

// 		for {
// 			// loop command
// 			selected, _, err := prompt.Run()

// 			// unknow error, should retry
// 			if err != nil {
// 				demo.LogInfo("Prompt failed %v\n", err)
// 				continue
// 			}

// 			// get selected command and run it
// 			command := commands[selected]
// 			if command.Name == "quit" {
// 				demo.LogInfo("Server quiting...")
// 				// commit changes to orderbook
// 				orderbookEngine.Commit()
// 				endWaiter.Done()
// 				thisNode.Stop()
// 				quitC <- struct{}{}
// 				demo.LogInfo("-> Goodbye\n")
// 				return
// 			}
// 			results := command.Run()

// 			// process command
// 			switch command.Name {
// 			case "processOrder":
// 				demo.LogInfo("-> Add order", "payload", results)
// 				// put message on channel
// 				go processOrder(results)

// 			// case "addNode":
// 			// 	nodeaddr := results["nodeaddr"]
// 			// 	demo.LogInfo(fmt.Sprintf("-> Add node: %s\n", nodeaddr))
// 			// 	addNode(nodeaddr)

// 			case "nodeAddr":
// 				demo.LogInfo(fmt.Sprintf("-> Node Address: %s\n", nodeAddr()))

// 			default:
// 				demo.LogInfo(fmt.Sprintf("-> Unknown command: %s\n", command.Name))
// 			}
// 		}

// 	}()

// 	// wait for command processing
// 	endWaiter.Wait()

// 	// finally shutdown
// 	return shutdown()
// }

// func shutdown() error {
// 	// return os.RemoveAll(thisNode.DataDir())
// 	return nil
// }

// // func addNode(rawurl string) error {

// // 	newNode, err := discover.ParseNode(rawurl)
// // 	if err != nil {
// // 		demo.LogCrit("pass node addr fail", "err", err)
// // 		return err
// // 	}

// // 	demo.LogInfo("add node", "node", newNode.String())
// // 	thisNode.Server().AddPeer(newNode)

// // 	// if have protocol implemented
// // 	if len(pssprotos) > 0 {
// // 		nid := newNode.ID
// // 		p := p2p.NewPeer(nid, nid.String(), []p2p.Cap{})

// // 		// calculate keys
// // 		// pubkey, _ := newNode.ID.Pubkey()
// // 		// pubkeyBytes := crypto.FromECDSAPub(pubkey)
// // 		// // ToHex will append 0x at first byte
// // 		// publicKey := common.ToHex(pubkeyBytes)
// // 		// bzzAddr := crypto.Keccak256Hash(pubkeyBytes).Hex()

// // 		// // fmt.Println("bzzAddr", bzzAddr, "public key", publicKey)
// // 		// // bzzaddr = "0x9984c9556ca87842c4ceb839518cd3648dc495d579f7af7f9ba49989bc207346"
// // 		// // get rpcClient
// // 		// rpcClient, _ := thisNode.Attach()

// // 		// // Set Public key to associate with a particular Pss peer
// // 		// err = rpcClient.Call(nil, "pss_setPeerPublicKey", publicKey, protocol.OrderbookTopic, bzzAddr)
// // 		// if err != nil {
// // 		// 	demo.LogCrit("pss set pubkey fail", "err", err)
// // 		// 	return err
// // 		// }

// // 		// add peer with its public key to this protocol on a topic and using asymetric cryptography
// // 		_, err = pssprotos[0].AddPeer(p, protocol.OrderbookTopic, true, publicKey)
// // 		if err != nil {
// // 			return err
// // 		}

// // 	}

// // 	demo.LogInfo("Added node successfully!")
// // 	return nil
// // }

// func nodeAddr() string {
// 	return thisNode.Server().Self().String()
// }

// func processOrder(payload map[string]string) error {
// 	// add order at this current node first
// 	// get timestamp in milliseconds
// 	payload["timestamp"] = strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
// 	msg, err := protocol.NewOrderbookMsg(payload)
// 	// if err == nil {
// 	// 	// try to store into model, if success then process at local and broad cast
// 	// 	// trades, orderInBook := orderbookEngine.ProcessOrder(payload)
// 	// 	// demo.LogInfo("Orderbook result", "Trade", trades, "OrderInBook", orderInBook)

// 	// 	// broad cast message
// 	// 	msgC <- msg

// 	// }
// 	// Set Public key to associate with a particular Pss peer
// 	rpcClient, _ := thisNode.Attach()
// 	// err = rpcClient.Call(nil, "pss_setPeerPublicKey", publicKey, protocol.OrderbookTopic, bzzAddr)
// 	// if err != nil {
// 	// 	demo.LogCrit("pss set pubkey fail", "err", err)
// 	// }
// 	raw, _ := rlp.EncodeToBytes(msg)
// 	err = rpcClient.Call(nil, "pss_sendRaw", bzzAddr, protocol.OrderbookTopic, raw)
// 	if err != nil {
// 		demo.LogCrit("Pss send fail", "err", err)
// 	}

// 	return nil
// }

// func updateEthService() {
// 	// make full node after start the node, then later register swarm service over that node
// 	ethConfig, err := initGenesis(thisNode)
// 	ethConfig.Etherbase = crypto.PubkeyToAddress(privkey.PublicKey)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	// register ethservice with genesis block
// 	utils.RegisterEthService(thisNode, ethConfig)

// 	if err != nil {
// 		demo.LogCrit("servicenode pss register fail", "err", err)
// 	}

// 	// config ethereum
// 	var ethereum *eth.Ethereum
// 	if err := thisNode.Service(&ethereum); err != nil {
// 		demo.LogError(fmt.Sprintf("Ethereum service not running: %v", err))
// 	}

// 	// config ethereum gas price
// 	ethereum.TxPool().SetGasPrice(ethConfig.GasPrice)

// 	password := "123456789"
// 	ks := thisNode.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
// 	// import will create a keystore if not found
// 	account, err := ks.ImportECDSA(privkey, password)
// 	err = ks.Unlock(account, password)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// if this node can mine
// 	if mining {
// 		miner := ethereum.Miner()
// 		miner.Start(account.Address)
// 	}
// }

// func updateExtraService(rpcapi []string) {

// 	for _, api := range rpcapi {
// 		if api == "eth" {
// 			updateEthService()
// 		}
// 	}

// }

// // simple ping and receive protocol
// func startup(p2pPort int, httpPort int, wsPort int, bzzPort int, privateKey string) {

// 	var err error

// 	// get private key
// 	privkey, err = crypto.LoadECDSA(privateKey)

// 	// register pss and orderbook service
// 	rpcapi := []string{
// 		// "eth",
// 		"bzz",
// 		"pss",
// 		"orderbook",
// 	}
// 	dataDir := fmt.Sprintf("%s%d", demo.DatadirPrefix, p2pPort)
// 	orderbookDir := path.Join(dataDir, "orderbook")
// 	allowedPairs := map[string]*big.Int{
// 		"TOMO/WETH": big.NewInt(10e9),
// 	}
// 	orderbookEngine = orderbook.NewEngine(orderbookDir, allowedPairs)

// 	thisNode, err = demo.NewServiceNodeWithPrivateKeyAndDataDir(privkey, dataDir, p2pPort, httpPort, wsPort, rpcapi...)

// 	// protocolSpecs := []*protocols.Spec{protocol.OrderbookProtocol}
// 	// proto := protocol.NewProtocol(msgC, quitC, orderbookEngine)
// 	// protocolArr := []*p2p.Protocol{proto}

// 	// register normal service, protocol is for p2p, service is for rpc calls
// 	// service := protocol.NewService(orderbookEngine)
// 	// err = thisNode.Register(service)

// 	// if err != nil {
// 	// 	demo.LogCrit("Register orderbook service in servicenode failed", "err", err)
// 	// }

// 	// register the pss activated bzz services, put the protocols into this service
// 	pssService := demo.NewSwarmServiceWithProtocolAndPrivateKey(thisNode, bzzPort, nil, nil, nil, nil, privkey)
// 	err = thisNode.Register(pssService)

// 	if err != nil {
// 		demo.LogCrit("servicenode pss register fail", "err", err)
// 	}

// 	// extra service like eth, swarm need more work
// 	updateExtraService(rpcapi)

// 	// start the nodes
// 	err = thisNode.Start()
// 	if err != nil {
// 		demo.LogCrit("servicenode start failed", "err", err)
// 	}

// 	newNode, _ := discover.ParseNode(nodeaddr)
// 	pubkey, _ := newNode.ID.Pubkey()
// 	pubkeyBytes := crypto.FromECDSAPub(pubkey)
// 	// ToHex will append 0x at first byte
// 	// publicKey = common.ToHex(pubkeyBytes)
// 	bzzAddr = crypto.Keccak256Hash(pubkeyBytes).Hex()
// 	demo.LogInfo("add node", "node", newNode.String())
// 	// if bzzPort == 8542 {
// 	thisNode.Server().AddPeer(newNode)
// 	// }
// 	rpcClient, _ := thisNode.Attach()
// 	sub, _ := rpcClient.Subscribe(context.Background(), "pss", msgC, "receive", protocol.OrderbookTopic)

// 	go func() {
// 		for {
// 			select {
// 			case payload := <-msgC:
// 				// demo.LogInfo("Internal received", "payload", payload)
// 				inmsg := &protocol.OrderbookMsg{}
// 				// maybe we have to use map[]chan
// 				rlp.DecodeBytes(payload.Msg, inmsg)
// 				// databytes, err := json.Marshal(inmsg)

// 				demo.LogDebug("Received orderbook", "orderbook", inmsg)

// 			// send quit command, break this loop
// 			case <-quitC:
// 				sub.Unsubscribe()
// 				break
// 			}
// 		}
// 	}()

// }

// // geth init genesis.json --datadir .datadir
// func initGenesis(stack *node.Node) (*eth.Config, error) {
// 	ethConfig := &eth.DefaultConfig
// 	ethConfig.SyncMode = downloader.FullSync
// 	ethConfig.SkipBcVersionCheck = true
// 	ethConfig.NetworkId = 8888

// 	var genesis core.Genesis
// 	databytes, err := ioutil.ReadFile(genesisPath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = json.Unmarshal(databytes, &genesis)
// 	ethConfig.Genesis = &genesis

// 	demo.LogInfo("Genesis", "chainid", genesis.Config.ChainId.String())
// 	return ethConfig, err
// }
