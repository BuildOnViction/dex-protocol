package main

import (
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tomochain/backend-matching-engine/utils/math"

	"github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb"
	"gopkg.in/urfave/cli.v1"
)

const wethABI = `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"},{"name":"","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"payable":true,"stateMutability":"payable","type":"fallback"},{"anonymous":false,"inputs":[{"indexed":true,"name":"src","type":"address"},{"indexed":true,"name":"guy","type":"address"},{"indexed":false,"name":"wad","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"src","type":"address"},{"indexed":true,"name":"dst","type":"address"},{"indexed":false,"name":"wad","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"dst","type":"address"},{"indexed":false,"name":"wad","type":"uint256"}],"name":"Deposit","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"src","type":"address"},{"indexed":false,"name":"wad","type":"uint256"}],"name":"Withdrawal","type":"event"},{"constant":false,"inputs":[],"name":"deposit","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"name":"wad","type":"uint256"}],"name":"withdraw","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"guy","type":"address"},{"name":"wad","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"dst","type":"address"},{"name":"wad","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"src","type":"address"},{"name":"dst","type":"address"},{"name":"wad","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

var mapping = map[string]uint64{
	"name":      0,
	"symbol":    1,
	"decimals":  2,
	"balanceOf": 3,
}

var (
	app             = cli.NewApp()
	contractAddress = common.HexToAddress("0xd645C13C35141d61f273EDc0F546beF48a48001D")
	user1Address    = common.HexToAddress("0x28074f8D0fD78629CD59290Cac185611a8d60109")
	user2Address    = common.HexToAddress("0x6e6BB166F420DDd682cAEbf55dAfBaFda74f2c9c")
	parsed          abi.ABI
	chaindb         *ethdb.LDBDatabase
	statedb         *state.StateDB
	blockNumber     uint64
)

// geth init genesis.json --datadir .datadir

func init() {
	parsed, _ = abi.JSON(strings.NewReader(wethABI))
	// privkey, _ := crypto.HexToECDSA("3411b45169aa5a8312e51357db68621031020dcf46011d7431db1bbb6d3922ce")
	// datadir := ".data_30100"
	// stack, _ := demo.NewServiceNodeWithPrivateKeyAndDataDir(privkey, dataDir, 0, 0, 0)
	datadir := "datadir/tomo/chaindata"
	// datadir := ".data_30100/main/chaindata"

	chaindb, _ = ethdb.NewLDBDatabase(datadir, 0, 0)

	// headHash := rawdb.ReadHeadBlockHash(chaindb)
	// blockNumber := *rawdb.ReadHeaderNumber(chaindb, headHash)

	headHash := core.GetHeadBlockHash(chaindb)
	blockNumber = core.GetBlockNumber(chaindb, headHash)

	// Initialize the CLI app and start tomo
	app.Commands = []cli.Command{
		cli.Command{
			Name: "save",
			Action: func(c *cli.Context) error {
				// must return export function
				return updateContract(c.String("method"), c.StringSlice("args"))
			},
			Flags: []cli.Flag{
				cli.StringFlag{Name: "method, m", Value: "symbol"},
				cli.StringSliceFlag{Name: "args, a"},
			},
		},
		cli.Command{
			Name: "load",
			Action: func(c *cli.Context) error {
				// must return export function
				return debugContract()
			},
		},
	}
}

func main() {

	// headHash := rawdb.ReadCanonicalHash(chaindb, number)
	// headerNumber := rawdb.ReadHeaderNumber(chaindb, headHash)
	// block := rawdb.ReadBlock(chaindb, headHash, *headerNumber)

	headHash := core.GetCanonicalHash(chaindb, blockNumber)
	// currentHeader = core.GetHeader(chaindb, headHash, number)
	blockNumber := core.GetBlockNumber(chaindb, headHash)
	block := core.GetBlock(chaindb, headHash, blockNumber)

	if block == nil {
		return
	}

	database := state.NewDatabase(chaindb)

	headerRootHash := block.Header().Root
	headHeaderHash := core.GetHeadHeaderHash(chaindb)
	statedb, _ = state.New(headHeaderHash, database)
	if statedb == nil {
		headHeaderHash = headerRootHash
	}

	fmt.Printf("Block number :%d, header root:%v\n", blockNumber, headerRootHash.Hex())

	statedb, _ = state.New(headHeaderHash, database)

	// utils.PrintJSON(block)

	if statedb != nil {
		if err := app.Run(os.Args); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

}

func debug(chaindb *ethdb.LDBDatabase, number uint64) {

}

func getKeyStorage(statedb *state.StateDB, address common.Address, methodName string, input ...common.Hash) (key common.Hash) {
	method, ok := parsed.Methods[methodName]
	slot := new(big.Int).SetUint64(mapping[methodName])
	// do not support function call
	if ok && len(method.Inputs) <= 1 || len(method.Outputs) == 1 {
		if len(method.Inputs) == 0 {
			key = getKey(slot)
		} else {
			// support first input

			key = getLocMap(slot, input[0])
		}
	}

	return key
}

func getStorage(statedb *state.StateDB, address common.Address, methodName string, input ...common.Hash) (key common.Hash, value interface{}) {
	key = getKeyStorage(statedb, address, methodName, input...)
	// ret := statedb.GetCommittedState(address, key)
	ret := statedb.GetState(address, key)

	method := parsed.Methods[methodName]
	switch method.Outputs[0].Type.T {
	case abi.StringTy:
		value = string(ret.Bytes())
	default:
		parsed.Unpack(&value, methodName, ret.Bytes())
	}

	// value = ret.Hex()
	return key, value
}

func setStorage(statedb *state.StateDB, address common.Address, methodName string, input ...common.Hash) (key, value common.Hash) {
	key = getKeyStorage(statedb, address, methodName, input...)
	value = input[len(input)-1]
	statedb.SetState(address, key, value)

	return key, value
}

func debugContract() error {

	// dump := statedb.RawDump()
	// dumpAcc := dump.Accounts[common.Bytes2Hex(address.Bytes())]
	// utils.PrintJSON(dumpAcc)

	// balance := new(big.Int).SetUint64(10000000)
	// statedb.AddBalance(user1Address, balance)
	// fmt.Printf("User: %v, balance: %v\n", user1Address.Hex(), statedb.GetBalance(user1Address).String())

	for _, methodName := range []string{"name", "symbol", "decimals"} {
		key, value := getStorage(statedb, contractAddress, methodName)
		fmt.Printf("key: %s, value :%v\n", key.Hex(), value)
	}

	for _, userAddress := range []common.Address{user1Address, user2Address} {
		key, value := getStorage(statedb, contractAddress, "balanceOf", userAddress.Hash())
		fmt.Printf("key: %s, value :%v\n", key.Hex(), value)
	}

	return nil
}

func updateContract(methodName string, args []string) error {
	// update
	// utils.PrintJSON(methodName)
	// utils.PrintJSON(args)
	// return nil

	var input []common.Hash
	for _, arg := range args {

		if arg[0:2] == "0x" {
			input = append(input, common.HexToHash(arg))
		} else {
			input = append(input, common.BytesToHash([]byte(arg)))
		}

	}

	// utils.PrintJSON(input)

	key, value := setStorage(statedb, contractAddress, methodName, input...)
	fmt.Printf("Update storage, key: %s, value :%v\n", key.Hex(), value.Hex())

	// commit
	root := statedb.IntermediateRoot(false)
	statedb.Commit(false)
	statedb.Database().TrieDB().Commit(root, true)
	fmt.Printf("New header root :%v\n", root.Hex())
	return core.WriteHeadHeaderHash(chaindb, root)

}

func getKey(slot *big.Int) common.Hash {
	return common.BigToHash(slot)
}

func getLocMap(slot *big.Int, key common.Hash) common.Hash {
	slotKey := getKey(slot)
	updatedKey := crypto.Keccak256Hash(key.Bytes(), slotKey.Bytes())
	return updatedKey
}

func getLocArr(slot, index, elementSize *big.Int) common.Hash {
	slotKey := getKey(slot)
	slotArr := new(big.Int).SetBytes(crypto.Keccak256Hash(slotKey.Bytes()).Bytes())
	slotArr = math.Add(slotArr, math.Mul(index, elementSize))
	return common.BigToHash(slotArr)
}
