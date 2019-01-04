// Copyright (c) 2015, Emir Pasic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package redblacktree implements a red-black tree.
//
// Used by TreeSet and TreeMap.
//
// Structure is not thread safe.
//
// References: http://en.wikipedia.org/wiki/Red%E2%80%93black_tree
package redblacktree

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
)

type color bool

type Comparator func(a, b []byte) int

const (
	black, red color = true, false
)

// Tree holds elements of the red-black tree
type Tree struct {
	db   *ethdb.LDBDatabase
	Root *Node
	// size       int
	Comparator Comparator
}

// NewWith instantiates a red-black tree with the custom comparator.
func NewWith(comparator Comparator, db *ethdb.LDBDatabase) *Tree {
	return &Tree{Comparator: comparator, db: db}
}

func NewWithBytesComparator(db *ethdb.LDBDatabase) *Tree {
	return &Tree{Comparator: bytes.Compare, db: db}
}

func (tree *Tree) SetRoot(key []byte) error {
	root, err := tree.GetNode(key)
	tree.Root = root
	// bytes, _ := json.Marshal(root)
	// fmt.Println(string(bytes))
	return err
}

// Put inserts node into the tree.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree) Put(key []byte, value []byte) {
	var insertedNode *Node
	if tree.Root == nil {
		// Assert key is of comparator's type for initial tree
		// tree.Comparator(key, key)
		item := &Item{Value: value, Color: red, Keys: &KeyMeta{}}
		tree.Root = &Node{Key: key, Item: item}
		insertedNode = tree.Root
	} else {
		node := tree.Root
		loop := true
		for loop {
			compare := tree.Comparator(key, node.Key)
			// fmt.Printf("Comparing :%v\n", compare)
			switch {
			case compare == 0:
				node.Key = key
				item := &Item{Value: value, Keys: &KeyMeta{}}
				node.Item = item
				return
			case compare < 0:
				if tree.emptyKey(node.LeftKey()) {
					node.LeftKey(key)
					item := &Item{Value: value, Color: red, Keys: &KeyMeta{}}
					nodeLeft := &Node{Key: key, Item: item}
					insertedNode = nodeLeft
					loop = false
				} else {
					node = node.Left(tree)
				}
			case compare > 0:

				if tree.emptyKey(node.RightKey()) {
					node.RightKey(key)
					item := &Item{Value: value, Color: red, Keys: &KeyMeta{}}
					nodeRight := &Node{Key: key, Item: item}
					insertedNode = nodeRight
					loop = false
				} else {
					// fmt.Printf("Noderight :%s:%s\n", node.RightKey(), key)
					node = node.Right(tree)
				}

			}
		}

		insertedNode.ParentKey(node.Key)
		tree.Save(node)
		tree.Save(tree.Root)
		// tree.Save(insertedNode)
		// fmt.Printf("Key :%s %s\n", node, insertedNode)
	}

	tree.insertCase1(insertedNode)

	// tree.size++
}

func (tree *Tree) GetNode(key []byte) (*Node, error) {
	if tree.emptyKey(key) {
		return nil, nil
	}
	// fmt.Printf("key: %s\n", string(key))
	bytes, err := tree.db.Get(key)
	if err != nil {
		return nil, err
	}
	item := &Item{}

	err = rlp.DecodeBytes(bytes, item)
	// err = json.Unmarshal(bytes, item)

	return &Node{Key: key, Item: item}, err
}

// Get searches the node in the tree by key and returns its value or nil if key is not found in tree.
// Second return parameter is true if key was found, otherwise false.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree) Get(key []byte) (value []byte, found bool) {
	node := tree.lookup(key)
	if node != nil {
		return node.Item.Value, true
	}
	return nil, false
}

// Remove remove the node from the tree by key.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree) Remove(key []byte) {
	var child *Node
	node := tree.lookup(key)
	if node == nil {
		return
	}
	var left, right *Node = nil, nil
	if !tree.emptyKey(node.LeftKey()) {
		left = node.Left(tree)
	}
	if !tree.emptyKey(node.RightKey()) {
		right = node.Right(tree)
	}

	if left != nil && right != nil {
		pred := left.maximumNode(tree)
		node.Key = pred.Key
		node.Item = pred.Item
		node = pred
	}
	if left == nil || right == nil {
		if right == nil {
			child = left
		} else {
			child = right
		}
		if node.Item.Color == black {
			node.Item.Color = nodeColor(child)
			tree.deleteCase1(node)
		}
		tree.replaceNode(node, child)
		if tree.emptyKey(node.ParentKey()) && child != nil {
			child.Item.Color = black
		}
	}
	// tree.size--
}

// // Empty returns true if tree does not contain any nodes
// func (tree *Tree) Empty() bool {
// 	return tree.size == 0
// }

// // Size returns number of nodes in the tree.
// func (tree *Tree) Size() int {
// 	return tree.size
// }

// // Keys returns all keys in-order
// func (tree *Tree) Keys() [][]byte {
// 	keys := make([][]byte, tree.size)
// 	it := tree.Iterator()
// 	for i := 0; it.Next(); i++ {
// 		keys[i] = it.Key()
// 	}
// 	return keys
// }

// // Values returns all values in-order based on the key.
// func (tree *Tree) Values() [][]byte {
// 	values := make([][]byte, tree.size)
// 	it := tree.Iterator()
// 	for i := 0; it.Next(); i++ {
// 		values[i] = it.Value()
// 	}
// 	return values
// }

// Left returns the left-most (min) node or nil if tree is empty.
func (tree *Tree) Left() *Node {
	var parent *Node
	current := tree.Root
	for current != nil {
		parent = current
		current = current.Left(tree)
	}
	return parent
}

// Right returns the right-most (max) node or nil if tree is empty.
func (tree *Tree) Right() *Node {
	var parent *Node
	current := tree.Root
	for current != nil {
		parent = current
		current = current.Right(tree)
	}
	return parent
}

// Floor Finds floor node of the input key, return the floor node or nil if no floor is found.
// Second return parameter is true if floor was found, otherwise false.
//
// Floor node is defined as the largest node that is smaller than or equal to the given node.
// A floor node may not be found, either because the tree is empty, or because
// all nodes in the tree are larger than the given node.
//
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree) Floor(key []byte) (floor *Node, found bool) {
	found = false
	node := tree.Root
	for node != nil {
		compare := tree.Comparator(key, node.Key)
		switch {
		case compare == 0:
			return node, true
		case compare < 0:
			node = node.Left(tree)
		case compare > 0:
			floor, found = node, true
			node = node.Right(tree)
		}
	}
	if found {
		return floor, true
	}
	return nil, false
}

// Ceiling finds ceiling node of the input key, return the ceiling node or nil if no ceiling is found.
// Second return parameter is true if ceiling was found, otherwise false.
//
// Ceiling node is defined as the smallest node that is larger than or equal to the given node.
// A ceiling node may not be found, either because the tree is empty, or because
// all nodes in the tree are smaller than the given node.
//
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree) Ceiling(key []byte) (ceiling *Node, found bool) {
	found = false
	node := tree.Root
	for node != nil {
		compare := tree.Comparator(key, node.Key)
		switch {
		case compare == 0:
			return node, true
		case compare < 0:
			ceiling, found = node, true
			node = node.Left(tree)
		case compare > 0:
			node = node.Right(tree)
		}
	}
	if found {
		return ceiling, true
	}
	return nil, false
}

// // Clear removes all nodes from the tree.
// func (tree *Tree) Clear() {
// 	tree.Root = nil
// 	tree.size = 0
// }

// String returns a string representation of container
func (tree *Tree) String() string {
	str := "RedBlackTree\n"

	// if !tree.Empty() {
	output(tree, tree.Root, "", true, &str)
	// }
	return str
}

func output(tree *Tree, node *Node, prefix string, isTail bool, str *string) {
	// fmt.Printf("Node : %v+\n", node)
	if !tree.emptyKey(node.RightKey()) {
		newPrefix := prefix
		if isTail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		output(tree, node.Right(tree), newPrefix, false, str)
	}
	*str += prefix
	if isTail {
		*str += "└── "
	} else {
		*str += "┌── "
	}
	*str += node.String() + "\n"
	if !tree.emptyKey(node.LeftKey()) {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		output(tree, node.Left(tree), newPrefix, true, str)
	}
}

func (tree *Tree) lookup(key []byte) *Node {
	node, _ := tree.GetNode(key)
	return node
	// node := tree.Root
	// for node != nil {
	// 	compare := tree.Comparator(key, node.Key)
	// 	switch {
	// 	case compare == 0:
	// 		return node
	// 	case compare < 0:
	// 		node = node.Left(tree)
	// 	case compare > 0:
	// 		node = node.Right(tree)
	// 	}
	// }
	// return nil
}

func (tree *Tree) rotateLeft(node *Node) {
	right := node.Right(tree)
	tree.replaceNode(node, right)
	// fmt.Printf("rotate: right: %s, left: %s\n", node.RightKey(), right.LeftKey())
	node.RightKey(right.LeftKey())
	if !tree.emptyKey(right.LeftKey()) {
		right.Left(tree).ParentKey(node.Key)
	}
	right.LeftKey(node.Key)
	// tree.Save(right)
	node.ParentKey(right.Key)
}

func (tree *Tree) rotateRight(node *Node) {
	left := node.Left(tree)
	tree.replaceNode(node, left)
	node.LeftKey(left.RightKey())
	if !tree.emptyKey(left.RightKey()) {
		left.Right(tree).ParentKey(node.Key)
	}
	left.RightKey(node.Key)
	// tree.Save(left)
	node.ParentKey(left.Key)
}

func (tree *Tree) replaceNode(old *Node, new *Node) {
	if tree.emptyKey(old.ParentKey()) {
		tree.Root = new
	} else {
		parent := old.Parent(tree)
		fmt.Printf("Update leftkey :%v, %v, %v\n", string(new.Key), string(old.Key), string(parent.Key))
		if tree.Comparator(old.Key, parent.LeftKey()) == 0 {

			parent.LeftKey(new.Key)
		} else {
			parent.RightKey(new.Key)
		}
	}
	if new != nil {
		new.ParentKey(old.ParentKey())
	}
}

func (tree *Tree) insertCase1(node *Node) {

	if tree.emptyKey(node.ParentKey()) {
		node.Item.Color = black
		// store this

		tree.Save(node)
	} else {

		tree.insertCase2(node)
	}
}

func (tree *Tree) insertCase2(node *Node) {
	if nodeColor(node.Parent(tree)) == black {
		tree.Save(node)
		return
	}

	tree.insertCase3(node)
}

func (tree *Tree) insertCase3(node *Node) {
	uncle := node.uncle(tree)
	parent := node.Parent(tree)
	grandparent := node.grandparent(tree)
	if nodeColor(uncle) == red {
		parent.Item.Color = black
		uncle.Item.Color = black
		tree.Save(uncle)
		grandparent.Item.Color = red
		tree.insertCase1(grandparent)
	} else {
		tree.insertCase4(node, parent, grandparent)
	}
}

func (tree *Tree) insertCase4(node, parent, grandparent *Node) {
	// grandparent := node.grandparent(tree)
	// parent := node.Parent(tree)
	if tree.Comparator(node.Key, parent.RightKey()) == 0 &&
		tree.Comparator(parent.Key, grandparent.LeftKey()) == 0 {
		tree.rotateLeft(parent)
		node = node.Left(tree)
	} else if tree.Comparator(node.Key, parent.LeftKey()) == 0 &&
		tree.Comparator(parent.Key, grandparent.RightKey()) == 0 {
		tree.rotateRight(parent)
		node = node.Right(tree)
	}

	tree.insertCase5(node, parent, grandparent)
}

func (tree *Tree) insertCase5(node, parent, grandparent *Node) {
	// parent := node.Parent(tree)
	parent.Item.Color = black
	// grandparent := node.grandparent(tree)
	grandparent.Item.Color = red
	if tree.Comparator(node.Key, parent.LeftKey()) == 0 &&
		tree.Comparator(parent.Key, grandparent.LeftKey()) == 0 {
		tree.rotateRight(grandparent)
	} else if tree.Comparator(node.Key, parent.RightKey()) == 0 &&
		tree.Comparator(parent.Key, grandparent.RightKey()) == 0 {

		// if string(node.Key) == "5" {
		// 	fmt.Println(tree)
		// }

		tree.rotateLeft(grandparent)

		// if string(node.Key) == "5" {
		// 	fmt.Println(tree)
		// }
	}
	tree.Save(parent)
	tree.Save(grandparent)

	// if string(node.Key) == "5" {
	// fmt.Println(node)
	// }

}

func (tree *Tree) Save(node *Node) error {
	// value, err := json.Marshal(node.Item)
	value, err := rlp.EncodeToBytes(node.Item)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return tree.db.Put(node.Key, value)
}

func (tree *Tree) emptyKey(key []byte) bool {
	return key == nil || len(key) == 0
}

func (tree *Tree) deleteCase1(node *Node) {
	if tree.emptyKey(node.ParentKey()) {
		return
	}
	tree.deleteCase2(node)
}

func (tree *Tree) deleteCase2(node *Node) {
	sibling := node.sibling(tree)
	if nodeColor(sibling) == red {
		parent := node.Parent(tree)
		parent.Item.Color = red
		sibling.Item.Color = black
		if tree.Comparator(node.Key, parent.LeftKey()) == 0 {
			tree.rotateLeft(parent)
		} else {
			tree.rotateRight(parent)
		}
	}
	tree.deleteCase3(node)
}

func (tree *Tree) deleteCase3(node *Node) {
	sibling := node.sibling(tree)
	parent := node.Parent(tree)
	if nodeColor(parent) == black &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.Left(tree)) == black &&
		nodeColor(sibling.Right(tree)) == black {
		sibling.Item.Color = red
		tree.deleteCase1(parent)
	} else {
		tree.deleteCase4(node)
	}
}

func (tree *Tree) deleteCase4(node *Node) {
	sibling := node.sibling(tree)
	parent := node.Parent(tree)
	if nodeColor(parent) == red &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.Left(tree)) == black &&
		nodeColor(sibling.Right(tree)) == black {
		sibling.Item.Color = red
		parent.Item.Color = black
	} else {
		tree.deleteCase5(node)
	}
}

func (tree *Tree) deleteCase5(node *Node) {
	sibling := node.sibling(tree)
	siblingLeft := sibling.Left(tree)
	siblingRight := sibling.Right(tree)
	parent := node.Parent(tree)

	if tree.Comparator(node.Key, parent.LeftKey()) == 0 &&
		nodeColor(sibling) == black &&
		nodeColor(siblingLeft) == red &&
		nodeColor(siblingRight) == black {
		sibling.Item.Color = red
		siblingLeft.Item.Color = black
		tree.rotateRight(sibling)
	} else if tree.Comparator(node.Key, parent.RightKey()) == 0 &&
		nodeColor(sibling) == black &&
		nodeColor(siblingRight) == red &&
		nodeColor(siblingLeft) == black {
		sibling.Item.Color = red
		siblingRight.Item.Color = black
		tree.rotateLeft(sibling)
	}
	tree.deleteCase6(node)
}

func (tree *Tree) deleteCase6(node *Node) {
	sibling := node.sibling(tree)
	siblingLeft := sibling.Left(tree)
	siblingRight := sibling.Right(tree)
	parent := node.Parent(tree)
	sibling.Item.Color = nodeColor(parent)
	parent.Item.Color = black
	if tree.Comparator(node.Key, parent.LeftKey()) == 0 && nodeColor(siblingRight) == red {
		siblingRight.Item.Color = black
		tree.rotateLeft(parent)
	} else if nodeColor(siblingLeft) == red {
		siblingLeft.Item.Color = black
		tree.rotateRight(parent)
	}

	// delete from db
	tree.db.Delete(node.Key)
}

func nodeColor(node *Node) color {
	if node == nil {
		return black
	}
	return node.Item.Color
}
