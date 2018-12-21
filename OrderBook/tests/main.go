package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
)

// TokenABI is the input ABI used to generate the binding from.
const TokenABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"mintingFinished\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"setOwner\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseApproval\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"finishMinting\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseApproval\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Mint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"MintFinished\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"SetOwner\",\"type\":\"event\"}]"

// const TokenBin = `0x60606040526004805460ff19169055341561001957600080fd5b604051604080610b74833981016040528080519190602001805160008054600160a060020a03191633600160a060020a031617905560035490925061006c9150826401000000006100c28102610a4b1704565b600355600160a060020a03821660009081526002602052604090205461009f9082640100000000610a4b6100c282021704565b600160a060020a03909216600090815260026020526040902091909155506100d8565b6000828201838110156100d157fe5b9392505050565b610a8d806100e76000396000f300606060405236156100cd5763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166305d2035b81146100d2578063095ea7b3146100f957806313af40351461011b57806318160ddd1461013c57806323b872dd1461016157806340c10f191461018957806366188463146101ab57806370a08231146101cd5780637d64bcb4146101ec5780638da5cb5b146101ff57806395d89b411461022e578063a9059cbb146102b8578063d73dd623146102da578063dd62ed3e146102fc575b600080fd5b34156100dd57600080fd5b6100e5610321565b604051901515815260200160405180910390f35b341561010457600080fd5b6100e5600160a060020a036004351660243561032a565b341561012657600080fd5b61013a600160a060020a0360043516610396565b005b341561014757600080fd5b61014f61041c565b60405190815260200160405180910390f35b341561016c57600080fd5b6100e5600160a060020a0360043581169060243516604435610422565b341561019457600080fd5b6100e5600160a060020a03600435166024356105a4565b34156101b657600080fd5b6100e5600160a060020a03600435166024356106a9565b34156101d857600080fd5b61014f600160a060020a03600435166107a3565b34156101f757600080fd5b6100e56107be565b341561020a57600080fd5b610212610829565b604051600160a060020a03909116815260200160405180910390f35b341561023957600080fd5b610241610838565b60405160208082528190810183818151815260200191508051906020019080838360005b8381101561027d578082015183820152602001610265565b50505050905090810190601f1680156102aa5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34156102c357600080fd5b6100e5600160a060020a036004351660243561086f565b34156102e557600080fd5b6100e5600160a060020a036004351660243561096a565b341561030757600080fd5b61014f600160a060020a0360043581169060243516610a0e565b60045460ff1681565b600160a060020a03338116600081815260016020908152604080832094871680845294909152808220859055909291907f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9259085905190815260200160405180910390a350600192915050565b60005433600160a060020a039081169116146103b157600080fd5b600054600160a060020a0380831691167fcbf985117192c8f614a58aaf97226bb80a754772f5f6edf06f87c675f2e6c66360405160405180910390a36000805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0392909216919091179055565b60035490565b6000600160a060020a038316151561043957600080fd5b600160a060020a03841660009081526002602052604090205482111561045e57600080fd5b600160a060020a038085166000908152600160209081526040808320339094168352929052205482111561049157600080fd5b600160a060020a0384166000908152600260205260409020546104ba908363ffffffff610a3916565b600160a060020a0380861660009081526002602052604080822093909355908516815220546104ef908363ffffffff610a4b16565b600160a060020a03808516600090815260026020908152604080832094909455878316825260018152838220339093168252919091522054610537908363ffffffff610a3916565b600160a060020a03808616600081815260016020908152604080832033861684529091529081902093909355908516917fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9085905190815260200160405180910390a35060019392505050565b6000805433600160a060020a039081169116146105c057600080fd5b60045460ff16156105d057600080fd5b6003546105e3908363ffffffff610a4b16565b600355600160a060020a03831660009081526002602052604090205461060f908363ffffffff610a4b16565b600160a060020a0384166000818152600260205260409081902092909255907f0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d41213968859084905190815260200160405180910390a2600160a060020a03831660007fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef8460405190815260200160405180910390a350600192915050565b600160a060020a0333811660009081526001602090815260408083209386168352929052908120548083111561070657600160a060020a03338116600090815260016020908152604080832093881683529290529081205561073d565b610716818463ffffffff610a3916565b600160a060020a033381166000908152600160209081526040808320938916835292905220555b600160a060020a0333811660008181526001602090815260408083209489168084529490915290819020547f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925915190815260200160405180910390a35060019392505050565b600160a060020a031660009081526002602052604090205490565b6000805433600160a060020a039081169116146107da57600080fd5b60045460ff16156107ea57600080fd5b6004805460ff191660011790557fae5184fba832cb2b1f702aca6117b8d265eaf03ad33eb133f19dde0f5920fa0860405160405180910390a150600190565b600054600160a060020a031681565b60408051908101604052600381527f544f4b0000000000000000000000000000000000000000000000000000000000602082015281565b6000600160a060020a038316151561088657600080fd5b600160a060020a0333166000908152600260205260409020548211156108ab57600080fd5b600160a060020a0333166000908152600260205260409020546108d4908363ffffffff610a3916565b600160a060020a033381166000908152600260205260408082209390935590851681522054610909908363ffffffff610a4b16565b600160a060020a0380851660008181526002602052604090819020939093559133909116907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9085905190815260200160405180910390a350600192915050565b600160a060020a0333811660009081526001602090815260408083209386168352929052908120546109a2908363ffffffff610a4b16565b600160a060020a0333811660008181526001602090815260408083209489168084529490915290819020849055919290917f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92591905190815260200160405180910390a350600192915050565b600160a060020a03918216600090815260016020908152604080832093909416825291909152205490565b600082821115610a4557fe5b50900390565b600082820183811015610a5a57fe5b93925050505600a165627a7a7230582069f8c037c76de7c279d025631584f6be90ca1ba57c2de246ace1b8c1f2dce76e0029`

var contractAddress = common.HexToAddress("0x620C38566BAD7a895cce707F42DCd5eaC1f94861")
var userAddress = common.HexToAddress("0x28074f8D0fD78629CD59290Cac185611a8d60109")

// geth init genesis.json --datadir .datadir
func initGenesis(genesisPath string) *core.Genesis {

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

type ContractAdmin struct {
}

func (c *ContractAdmin) Address() common.Address {
	return contractAddress
}

func DumpTrie(trieItem state.Trie) {
	it := trie.NewIterator(trieItem.NodeIterator(nil))
	for it.Next() {
		fmt.Printf("Key: %x, value: %x\n", it.Key, it.Value)
	}
}

func main() {

	// privkey, _ := crypto.HexToECDSA("3411b45169aa5a8312e51357db68621031020dcf46011d7431db1bbb6d3922ce")
	// dataDir := ".data_30100"
	// stack, _ := demo.NewServiceNodeWithPrivateKeyAndDataDir(privkey, dataDir, 0, 0, 0)

	chainDb, _ := ethdb.NewLDBDatabase("datadir/geth/chaindata", 0, 0)
	// chainDb, _ := stack.OpenDatabase("chaindata", 0, 0)

	// rootHash := rawdb.ReadCanonicalHash(chainDb, 0)

	// genesisPath := "OrderBook/genesis.json"
	// genesisConfig := initGenesis(genesisPath)

	// config, _, _ := core.SetupGenesisBlock(chainDb, genesisConfig)

	headHash := rawdb.ReadHeadBlockHash(chainDb)
	headerNumber := rawdb.ReadHeaderNumber(chainDb, headHash)
	block := rawdb.ReadBlock(chainDb, headHash, *headerNumber)

	// utils.PrintJSON(headerNumber)
	// engine := clique.New(config.Clique, chainDb)

	// vmcfg := vm.Config{EnablePreimageRecording: false}
	// chain, _ := core.NewBlockChain(chainDb, nil, config, nil, vmcfg, nil)

	// fmt.Printf("root : %s, current: %s\n", rootHash.String(), chain.CurrentBlock().Hash().String())

	// currentBlockNumber := uint64(chain.CurrentBlock().Number().Int64())
	// lastBlock := chain.GetBlockByNumber(currentBlockNumber)
	// rootKey := common.HexToHash("0x5f76e530b22d37e7edb414986b21a4077579e3e4ea2c6825e68ac4fb6c88b6c5")

	// client := rpc.DialInProc(stack.Server())

	// backend := ethclient.NewClient(client)
	// instance, _ := contractsinterfaces.NewToken(contractAddress, backend)
	// callOpts := &bind.CallOpts{Pending: true}
	// tokenBalance, _ := instance.BalanceOf(callOpts, userAddress)

	// fmt.Printf("Balance AE of node1: %s", tokenBalance.String())
	// statedb, _ := state.New(rootHash, database)
	// fmt.Print(database.TrieDB().Size())

	database := state.NewDatabase(chainDb)
	statedb, _ := state.New(block.Root(), database)
	// statedb, _ := chain.StateAt(block.Root())

	// statedb, _ := chain.State()

	// trie := statedb.StorageTrie(contractAddress)

	// balance := statedb.GetBalance(userAddress)
	// fmt.Printf("Balance :%s", balance.String())
	// statedb.SetBalance(userAddress, balance)
	dump := statedb.RawDump()

	contractData := dump.Accounts[common.Bytes2Hex(contractAddress.Bytes())]

	for key, value := range contractData.Storage {
		bytesValue := common.Hex2Bytes(value)
		var data interface{}
		rlp.DecodeBytes(bytesValue, &data)
		fmt.Printf("Key: %s, value: %v\n", key, data)
	}

	// utils.PrintJSON(contractData)

	// transactions := lastBlock.Transactions()
	// transactions := chain.GetBlockByNumber(7).Transactions()
	// utils.PrintJSON(currentBlockNumber)

	// utils.PrintJSON(transactions)

	// parsed, _ := abi.JSON(strings.NewReader(TokenABI))
	// utils.PrintJSON(parsed)

	// rawTx := types.NewTransaction(0, contractAddress, nil, 210000, nil, input)

	// auth := bind.NewKeyedTransactor(privkey)

	// receipt := chain.GetReceiptsByHash(currentTx[1].Hash())
	// blockNumber := new(big.Int)
	// blockNumber.SetUint64(currentBlockNumber)
	// evmContext := vm.Context{
	// 	CanTransfer: core.CanTransfer,
	// 	Transfer:    core.Transfer,
	// 	GetHash:     func(uint64) common.Hash { return common.Hash{} },

	// 	Coinbase:    userAddress,
	// 	BlockNumber: blockNumber,
	// 	GasLimit:    genesisConfig.GasLimit,
	// 	Time:        new(big.Int).SetUint64(genesisConfig.Timestamp),
	// 	Difficulty:  genesisConfig.Difficulty,
	// }
	// statedb.SetCode(contractAddress, common.FromHex(TokenBin))
	// vmenv := vm.NewEVM(evmContext, statedb, config, vmcfg)
	// contractAdmin := vm.AccountRef(userAddress)
	// output, _, _ := vmenv.StaticCall(contractAdmin, contractAddress, input, 210000)
	// code := statedb.GetCode(contractAddress)

	// fmt.Printf("input :%x\n", input)
	// fmt.Printf("contract code: %x", code)
	// var outputRet interface{}
	// parsed.Unpack(&outputRet, "balanceOf", output)

	// fmt.Printf("contract result: %x\n", output)
	// var i uint64
	// var prevKey common.Hash
	// for ; i < maxResult; i++ {
	// 	block := chain.GetBlockByNumber(i)

	// 	stateDb := state.NewDatabase(chainDb)

	// 	stateRoot := block.Header().Root

	// 	// trie0, err := stateDb.OpenStorageTrie(contractAddress, stateRoot)
	// 	trie0, err := stateDb.OpenTrie(stateRoot)

	// 	if err != nil {
	// 		// fmt.Print(err)
	// 		continue
	// 	}

	// }

}
