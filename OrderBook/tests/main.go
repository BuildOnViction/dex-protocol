package main

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb"
)

const wethABI = `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"},{"name":"","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"payable":true,"stateMutability":"payable","type":"fallback"},{"anonymous":false,"inputs":[{"indexed":true,"name":"src","type":"address"},{"indexed":true,"name":"guy","type":"address"},{"indexed":false,"name":"wad","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"src","type":"address"},{"indexed":true,"name":"dst","type":"address"},{"indexed":false,"name":"wad","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"dst","type":"address"},{"indexed":false,"name":"wad","type":"uint256"}],"name":"Deposit","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"src","type":"address"},{"indexed":false,"name":"wad","type":"uint256"}],"name":"Withdrawal","type":"event"},{"constant":false,"inputs":[],"name":"deposit","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"name":"wad","type":"uint256"}],"name":"withdraw","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"guy","type":"address"},{"name":"wad","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"dst","type":"address"},{"name":"wad","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"src","type":"address"},{"name":"dst","type":"address"},{"name":"wad","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

var mapping = map[string]uint64{
	"name":      0,
	"symbol":    1,
	"decimals":  2,
	"balanceOf": 3,
}

var contractAddress = common.HexToAddress("0xd645C13C35141d61f273EDc0F546beF48a48001D")
var user1Address = common.HexToAddress("0x28074f8D0fD78629CD59290Cac185611a8d60109")
var user2Address = common.HexToAddress("0x6e6BB166F420DDd682cAEbf55dAfBaFda74f2c9c")
var parsed abi.ABI

// geth init genesis.json --datadir .datadir

func init() {
	parsed, _ = abi.JSON(strings.NewReader(wethABI))
}

func main() {
	// privkey, _ := crypto.HexToECDSA("3411b45169aa5a8312e51357db68621031020dcf46011d7431db1bbb6d3922ce")
	// datadir := ".data_30100"
	// stack, _ := demo.NewServiceNodeWithPrivateKeyAndDataDir(privkey, dataDir, 0, 0, 0)
	datadir := "datadir/geth/chaindata"
	// datadir := ".data_30100/main/chaindata"
	chainDb, _ := ethdb.NewLDBDatabase(datadir, 0, 0)
	headHash := rawdb.ReadHeadBlockHash(chainDb)
	headerNumber := rawdb.ReadHeaderNumber(chainDb, headHash)
	debug(chainDb, *headerNumber)
}

func debug(chainDb *ethdb.LDBDatabase, number uint64) {

	headHash := rawdb.ReadCanonicalHash(chainDb, number)
	headerNumber := rawdb.ReadHeaderNumber(chainDb, headHash)
	block := rawdb.ReadBlock(chainDb, headHash, *headerNumber)

	database := state.NewDatabase(chainDb)
	statedb, _ := state.New(block.Root(), database)
	if statedb != nil {
		debugContract(statedb, contractAddress)
	}
}

func getKeyStorage(statedb *state.StateDB, address common.Address, methodName string, input ...common.Hash) (key common.Hash) {
	method, ok := parsed.Methods[methodName]
	slot := mapping[methodName]
	// do not support function call
	if ok && len(method.Inputs) <= 1 || len(method.Outputs) == 1 {
		if len(method.Inputs) == 0 {
			key = getKey(slot)
		} else {
			// support first input
			key = getKeyMap(input[0], slot)
		}
	}

	return key
}

func getStorage(statedb *state.StateDB, address common.Address, methodName string, input ...common.Hash) (key common.Hash, value interface{}) {
	key = getKeyStorage(statedb, address, methodName, input...)
	ret := statedb.GetCommittedState(address, key)
	method := parsed.Methods[methodName]
	switch method.Outputs[0].Type.T {
	case abi.StringTy:
		value = string(ret.Bytes())
	default:
		parsed.Unpack(&value, methodName, ret.Bytes())
	}

	return key, value
}

func setStorage(statedb *state.StateDB, address common.Address, methodName string, input ...common.Hash) (key, value common.Hash) {
	key = getKeyStorage(statedb, address, methodName, input...)
	value = input[len(input)-1]
	statedb.SetState(address, key, value)
	statedb.Commit(false)
	// return
	return key, value
}

func debugContract(statedb *state.StateDB, address common.Address) {

	newName := common.BytesToHash([]byte("TOMO"))
	key, value := setStorage(statedb, contractAddress, "symbol", newName)
	fmt.Printf("key: %s, value :%v\n", key.Hex(), value.Hex())

	for _, methodName := range []string{"name", "symbol", "decimals"} {
		key, value := getStorage(statedb, contractAddress, methodName)
		fmt.Printf("key: %s, value :%v\n", key.Hex(), value)
	}

	for _, userAddress := range []common.Address{user1Address, user2Address} {
		key, value := getStorage(statedb, contractAddress, "balanceOf", userAddress.Hash())
		fmt.Printf("key: %s, value :%v\n", key.Hex(), value)
	}

}

func getKey(slot uint64) common.Hash {
	updatedKey := common.BigToHash(new(big.Int).SetUint64(slot))
	return updatedKey
}

func getKeyMap(key common.Hash, slot uint64) common.Hash {
	slotKey := common.BigToHash(new(big.Int).SetUint64(slot))
	updatedKey := crypto.Keccak256Hash(key.Bytes(), slotKey.Bytes())
	return updatedKey
}
