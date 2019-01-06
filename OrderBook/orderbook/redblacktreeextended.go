// Copyright (c) 2019, Agiletech Viet Nam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package orderbook

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	rbt "github.com/tomochain/orderbook/redblacktree"
)

func RLPEncodeToBytes(item *rbt.Item) ([]byte, error) {
	return rlp.EncodeToBytes(item)
}

func RLPDecodeBytes(bytes []byte, item *rbt.Item) error {
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

func OffsetEncodeBytes(item *rbt.Item) ([]byte, error) {
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

func OffsetDecodeBytes(bytes []byte, item *rbt.Item) error {
	start := 3 * common.HashLength
	totalLength := len(bytes)
	if item.Keys == nil {
		item.Keys = &rbt.KeyMeta{
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

// RedBlackTreeExtended to demonstrate how to extend a RedBlackTree to include new functions
type RedBlackTreeExtended struct {
	*rbt.Tree
}

func NewRedBlackTreeExtended(datadir string) *RedBlackTreeExtended {

	// setup using bytes offset
	obdb, _ := ethdb.NewLDBDatabase(datadir, 128, 1024)
	emptyKey := make([]byte, common.HashLength)
	// tree := &RedBlackTreeExtended{NewWithBytesComparator(RLPEncodeToBytes, RLPDecodeBytes, obdb)}
	tree := &RedBlackTreeExtended{rbt.NewWith(CmpBigInt, OffsetEncodeBytes, OffsetDecodeBytes, emptyKey, obdb)}

	tree.FormatBytes = func(key []byte) string {
		if len(key) == 0 || key == nil {
			return "<nil>"
		}
		return new(big.Int).SetBytes(key).String()
	}

	return tree
}

// GetMin gets the min value and flag if found
func (tree *RedBlackTreeExtended) GetMin() (value []byte, found bool) {
	node, found := tree.getMinFromNode(tree.Root())
	if node != nil {
		return node.Value(), found
	}
	return nil, false
}

// GetMax gets the max value and flag if found
func (tree *RedBlackTreeExtended) GetMax() (value []byte, found bool) {
	node, found := tree.getMaxFromNode(tree.Root())
	if node != nil {
		return node.Value(), found
	}
	return nil, false
}

// RemoveMin removes the min value and flag if found
func (tree *RedBlackTreeExtended) RemoveMin() (value []byte, deleted bool) {
	node, found := tree.getMinFromNode(tree.Root())
	// fmt.Println("found min", node)
	if found {
		tree.Remove(node.Key)
		return node.Value(), found
	}
	return nil, false
}

// RemoveMax removes the max value and flag if found
func (tree *RedBlackTreeExtended) RemoveMax() (value []byte, deleted bool) {
	// fmt.Println("found max with root", tree.Root())
	node, found := tree.getMaxFromNode(tree.Root())
	// fmt.Println("found max", node)
	if found {
		tree.Remove(node.Key)
		return node.Value(), found
	}
	return nil, false
}

func (tree *RedBlackTreeExtended) getMinFromNode(node *rbt.Node) (foundNode *rbt.Node, found bool) {
	if node == nil {
		return nil, false
	}
	nodeLeft := node.Left(tree.Tree)
	if nodeLeft == nil {
		return node, true
	}
	return tree.getMinFromNode(nodeLeft)
}

func (tree *RedBlackTreeExtended) getMaxFromNode(node *rbt.Node) (foundNode *rbt.Node, found bool) {
	if node == nil {
		return nil, false
	}
	nodeRight := node.Right(tree.Tree)
	if nodeRight == nil {
		return node, true
	}
	return tree.getMaxFromNode(nodeRight)
}
