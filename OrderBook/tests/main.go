package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/ethereum/go-ethereum/consensus/clique"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/node"
	demo "github.com/tomochain/orderbook/common"
)

// geth init genesis.json --datadir .datadir
func initGenesis(stack *node.Node, genesisPath string) *core.Genesis {

	var genesis core.Genesis
	databytes, _ := ioutil.ReadFile(genesisPath)

	json.Unmarshal(databytes, &genesis)
	return &genesis
}

func printJSON(x interface{}) {
	b, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Print(string(b), "\n")
}

func main() {

	privkey, _ := crypto.HexToECDSA("3411b45169aa5a8312e51357db68621031020dcf46011d7431db1bbb6d3922ce")
	stack, _ := demo.NewServiceNodeWithPrivateKey(privkey, 0, 0, 0)

	chainDb, _ := stack.OpenDatabase("chaindata", 0, 0)
	genesisPath := "OrderBook/genesis.json"
	config, _, _ := core.SetupGenesisBlock(chainDb, initGenesis(stack, genesisPath))

	engine := clique.New(config.Clique, chainDb)

	vmcfg := vm.Config{EnablePreimageRecording: false}
	chain, _ := core.NewBlockChain(chainDb, nil, config, engine, vmcfg, nil)

	var maxResult = uint64(chain.CurrentBlock().Number().Int64())
	// contractAddress := common.HexToAddress("0x620C38566BAD7a895cce707F42DCd5eaC1f94861")
	// rootKey := common.HexToHash("0x5f76e530b22d37e7edb414986b21a4077579e3e4ea2c6825e68ac4fb6c88b6c5")
	// addressHash := common.HexToAddress("0x28074f8D0fD78629CD59290Cac185611a8d60109")
	// client := rpc.DialInProc(stack.Server())

	// backend := ethclient.NewClient(client)
	// instance, _ := contractsinterfaces.NewToken(contractAddress, backend)
	// callOpts := &bind.CallOpts{Pending: true}
	// tokenBalance, _ := instance.BalanceOf(callOpts, addressHash)

	// fmt.Printf("Balance AE of node1: %s", tokenBalance.String())

	var i uint64
	var prevKey common.Hash
	for ; i < maxResult; i++ {
		block := chain.GetBlockByNumber(i)

		stateDb := state.NewDatabase(chainDb)

		stateRoot := block.Header().Root

		// trie0, err := stateDb.OpenStorageTrie(contractAddress, stateRoot)
		trie0, err := stateDb.OpenTrie(stateRoot)
		if err != nil {
			// fmt.Print(err)
			continue
		}

		it := trie.NewIterator(trie0.NodeIterator(nil))

		var itIndex uint
		for it.Next() {

			key := common.BytesToHash(it.Key)
			if prevKey == key {
				continue
			}
			prevKey = key
			var account state.Account
			rlp.DecodeBytes(it.Value, &account)
			// time.Sleep(50 * time.Millisecond)

			if (account.Root != common.Hash{} && account.Balance.Int64() != 0) {
				fmt.Printf("Key: %s, block number: %d, interator: %d\n", key.Hex(), i, itIndex)
				printJSON(account)
			}

			itIndex++
		}

	}

}
