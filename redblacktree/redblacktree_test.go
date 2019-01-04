package redblacktree

import (
	"testing"

	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/tomochain/orderbook/orderbook"
)

func print(tree *RedBlackTreeExtended, t *testing.T) {
	max, _ := tree.GetMax()
	min, _ := tree.GetMin()
	t.Logf("Value for max key: %v \n", string(max))
	t.Logf("Value for min key: %v \n", string(min))
	t.Log(tree)
}

func printOrgin(tree *orderbook.RedBlackTreeExtended, t *testing.T) {
	max, _ := tree.GetMax()
	min, _ := tree.GetMin()
	t.Logf("Value for max key: %v \n", max)
	t.Logf("Value for min key: %v \n", min)
	t.Log(tree)
}

func TestLevelDB(t *testing.T) {
	// current running folder is this folder
	datadir := "../datadir/tomo/orderbook"
	obdb, _ := ethdb.NewLDBDatabase(datadir, 0, 0)
	// obdb.Put([]byte("1"), []byte("a"))
	value, _ := obdb.Get([]byte("1"))

	t.Logf("value :%s", string(value))
}

func TestManipulateOriginTree(t *testing.T) {

	tree := &orderbook.RedBlackTreeExtended{rbt.NewWithStringComparator()}

	tree.Put("1", "a") // 1->a (in order)
	tree.Put("2", "b") // 1->a, 2->b (in order)
	tree.Put("3", "c") // 1->a, 2->b, 3->c (in order)
	tree.Put("4", "d") // 1->a, 2->b, 3->c, 4->d (in order)
	tree.Put("5", "e") // 1->a, 2->b, 3->c, 4->d, 5->e (in order)

	printOrgin(tree, t)

}

func TestManipulateLevelDBTree(t *testing.T) {

	datadir := "../datadir/tomo/orderbook"
	obdb, _ := ethdb.NewLDBDatabase(datadir, 0, 0)

	tree := &RedBlackTreeExtended{NewWithBytesComparator(obdb)}

	tree.Put([]byte("1"), []byte("a")) // 1->a (in order)
	tree.Put([]byte("2"), []byte("b")) // 1->a, 2->b (in order)
	tree.Put([]byte("3"), []byte("c")) // 1->a, 2->b, 3->c (in order)
	tree.Put([]byte("4"), []byte("d")) // 1->a, 2->b, 3->c, 4->d (in order)
	tree.Put([]byte("5"), []byte("e")) // 1->a, 2->b, 3->c, 4->d, 5->e (in order)

	print(tree, t)
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

	// print(&tree, t)
	// Value for max key: c
	// Value for min key: c
	// RedBlackTree
	// └── 3
}

func TestRestoreLevelDBTree(t *testing.T) {
	datadir := "../datadir/tomo/orderbook"
	obdb, _ := ethdb.NewLDBDatabase(datadir, 0, 0)

	tree := &RedBlackTreeExtended{NewWithBytesComparator(obdb)}
	tree.SetRoot([]byte("2"))

	print(tree, t)
}
