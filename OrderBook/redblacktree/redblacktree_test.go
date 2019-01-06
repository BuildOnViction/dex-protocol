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

func getBig(value string) []byte {
	bigValue, _ := new(big.Int).SetString(value, 10)
	return common.BigToHash(bigValue).Bytes()
}

func TestManipulateLevelDBTree(t *testing.T) {
	datadir := "datadir/agiletech/orderbook"
	tree := NewRedBlackTreeExtended(datadir)

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
	tree := NewRedBlackTreeExtended(datadir)

	tree.SetRootKey(getBig("2"))

	print(tree, t)
}
