package redblacktree

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
)

func print(tree *RedBlackTreeExtended, t *testing.T) {
	max, _ := tree.GetMax()
	min, _ := tree.GetMin()
	t.Logf("Value for max key: %v \n", string(max))
	t.Logf("Value for min key: %v \n", string(min))
	t.Log(tree)
}

func TestLevelDB(t *testing.T) {
	// current running folder is this folder
	datadir := "datadir/agiletech/orderbook"
	obdb, _ := ethdb.NewLDBDatabase(datadir, 0, 0)
	// obdb.Put([]byte("1"), []byte("a"))
	value, _ := obdb.Get([]byte("2"))
	item := &Item{}
	rlp.DecodeBytes(value, item)
	t.Logf("value :%x, items : %s", value, item.Value)
}

func RLPEncodeToBytes(item *Item) ([]byte, error) {
	return rlp.EncodeToBytes(item)
}

func RLPDecodeBytes(bytes []byte, item *Item) error {
	return rlp.DecodeBytes(bytes, item)
}

const (
	trueByte  = byte(1)
	falseByte = byte(0)
)

func bool2byte(bln bool) byte {
	if bln == true {
		return trueByte
	}

	return falseByte
}

func byte2bool(b byte) bool {
	if b == trueByte {
		return true
	}
	return false
}

func OffsetEncodeBytes(item *Item) ([]byte, error) {
	start := 3 * common.HashLength
	totalLength := start + 2
	if item.Value != nil {
		totalLength += len(item.Value)
	}

	returnBytes := make([]byte, totalLength)

	if item.Keys != nil {
		copy(returnBytes[0:common.HashLength], item.Keys.Left)
		copy(returnBytes[common.HashLength:2*common.HashLength], item.Keys.Right)
		copy(returnBytes[2*common.HashLength:start], item.Keys.Parent)
	}
	returnBytes[start] = bool2byte(item.Color)
	start++
	returnBytes[start] = bool2byte(item.Deleted)
	start++
	if start < totalLength {
		copy(returnBytes[start:], item.Value)
	}

	// fmt.Printf("value :%x\n", returnBytes)

	return returnBytes, nil
}

func OffsetDecodeBytes(bytes []byte, item *Item) error {
	start := 3 * common.HashLength
	totalLength := len(bytes)
	if item.Keys == nil {
		item.Keys = &KeyMeta{
			Left:   make([]byte, common.HashLength),
			Right:  make([]byte, common.HashLength),
			Parent: make([]byte, common.HashLength),
		}
	}
	copy(item.Keys.Left, bytes[0:common.HashLength])
	copy(item.Keys.Right, bytes[common.HashLength:2*common.HashLength])
	copy(item.Keys.Parent, bytes[2*common.HashLength:start])
	item.Color = byte2bool(bytes[start])
	start++
	item.Deleted = byte2bool(bytes[start])
	start++
	if start < totalLength {
		item.Value = make([]byte, totalLength-start)
		copy(item.Value, bytes[start:])
	}

	// fmt.Printf("Item key : %#v\n", item.Keys)

	return nil
}

func NewTree(datadir string) *RedBlackTreeExtended {

	obdb, _ := ethdb.NewLDBDatabase(datadir, 128, 1024)
	emptyKey := make([]byte, common.HashLength)
	// tree := &RedBlackTreeExtended{NewWithBytesComparator(RLPEncodeToBytes, RLPDecodeBytes, obdb)}
	tree := &RedBlackTreeExtended{NewWithBytesComparator(OffsetEncodeBytes, OffsetDecodeBytes, emptyKey, obdb)}

	tree.FormatBytes = func(key []byte) string {
		if len(key) == 0 || key == nil {
			return "<nil>"
		}
		return new(big.Int).SetBytes(key).String()
	}

	return tree
}

func getBig(value string) []byte {
	bigValue, _ := new(big.Int).SetString(value, 10)
	return common.BigToHash(bigValue).Bytes()
}

func TestManipulateLevelDBTree(t *testing.T) {
	datadir := "datadir/agiletech/orderbook"
	tree := NewTree(datadir)

	start := time.Now()
	tree.Put(getBig("1"), []byte("a")) // 1->a (in order)
	tree.Put(getBig("2"), []byte("b")) // 1->a, 2->b (in order)
	tree.Put(getBig("3"), []byte("c")) // 1->a, 2->b, 3->c (in order)
	tree.Put(getBig("4"), []byte("d")) // 1->a, 2->b, 3->c, 4->d (in order)
	tree.Put(getBig("5"), []byte("e")) // 1->a, 2->b, 3->c, 4->d, 5->e (in order)

	t.Logf("Done operation took: %v", time.Since(start))

	print(tree, t)

	tree.Commit()
	// Value for max key: e
	// Value for min key: a
	// RedBlackTree
	// │       ┌── 5
	// │   ┌── 4
	// │   │   └── 3
	// └── 2
	//     └── 1

	// tree.RemoveMin() // 2->b, 3->c, 4->d, 5->e (in order)
	// tree.RemoveMax() // 2->b, 3->c, 4->d (in order)
	// tree.RemoveMin() // 3->c, 4->d (in order)
	// tree.RemoveMax() // 3->c (in order)

	// print(tree, t)
	// Value for max key: c
	// Value for min key: c
	// RedBlackTree
	// └── 3
}

func TestRestoreLevelDBTree(t *testing.T) {
	datadir := "datadir/agiletech/orderbook"
	tree := NewTree(datadir)

	tree.SetRootKey(getBig("2"))

	print(tree, t)
}
