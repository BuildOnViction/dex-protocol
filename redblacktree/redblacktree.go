// Copyright (c) 2019, Agiletech Viet Nam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// References: http://en.wikipedia.org/wiki/Red%E2%80%93black_tree
package redblacktree

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/ethdb"
)

type Comparator func(a, b []byte) int
type EncodeToBytes func(*Item) ([]byte, error)
type DecodeBytes func([]byte, *Item) error

var emptyKey = []byte{}

// Tree holds elements of the red-black tree
type Tree struct {
	db      *ethdb.LDBDatabase
	rootKey []byte
	// size       int
	Comparator    Comparator
	EncodeToBytes EncodeToBytes
	DecodeBytes   DecodeBytes
}

// NewWith instantiates a red-black tree with the custom comparator.
func NewWith(comparator Comparator, encode EncodeToBytes, decode DecodeBytes, db *ethdb.LDBDatabase) *Tree {
	return &Tree{Comparator: comparator, EncodeToBytes: encode, DecodeBytes: decode, db: db}
}

func NewWithBytesComparator(encode EncodeToBytes, decode DecodeBytes, db *ethdb.LDBDatabase) *Tree {
	return &Tree{Comparator: bytes.Compare, EncodeToBytes: encode, DecodeBytes: decode, db: db}
}

func (tree *Tree) Root() *Node {
	root, _ := tree.GetNode(tree.rootKey)
	return root
}

func (tree *Tree) SetRootKey(key []byte) {
	// root, err := tree.GetNode(key)
	// tree.Root = root
	// // bytes, _ := json.Marshal(root)
	// // fmt.Println(string(bytes))
	// return err
	tree.rootKey = key
}

// Put inserts node into the tree.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree) Put(key []byte, value []byte) {
	var insertedNode *Node
	if tree.emptyKey(tree.rootKey) {
		// Assert key is of comparator's type for initial tree
		// tree.Comparator(key, key)
		item := &Item{Value: value, Color: red, Keys: &KeyMeta{}}
		tree.rootKey = key
		insertedNode = &Node{Key: key, Item: item}
	} else {
		node := tree.Root()
		loop := true
		for loop {
			compare := tree.Comparator(key, node.Key)
			// fmt.Printf("Comparing :%v\n", compare)
			switch {
			case compare == 0:
				node.Key = key
				item := &Item{Value: value, Keys: &KeyMeta{}}
				node.Item = item
				tree.Save(node)
				return
			case compare < 0:
				if tree.emptyKey(node.LeftKey()) {
					node.LeftKey(key)
					tree.Save(node)
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
					tree.Save(node)
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
		tree.Save(insertedNode)

		// fmt.Printf("Key :%s %s\n", node, insertedNode)
	}

	// tree.Save(insertedNode)
	tree.insertCase1(insertedNode)
	tree.Save(insertedNode)

	fmt.Println(tree)
	// tree.size++
}

func (tree *Tree) GetNode(key []byte) (*Node, error) {
	if tree.emptyKey(key) {
		return nil, nil
	}
	// fmt.Printf("key: %s\n", string(key))
	bytes, err := tree.db.Get(key)
	if err != nil {
		fmt.Printf("Key not found :%s", string(key))
		return nil, err
	}
	item := &Item{}

	err = tree.DecodeBytes(bytes, item)
	// err = json.Unmarshal(bytes, item)

	// fmt.Printf("Bytes :%v", item)

	if item.Deleted {
		return nil, nil
	}
	return &Node{Key: key, Item: item}, err
}

// Get searches the node in the tree by key and returns its value or nil if key is not found in tree.
// Second return parameter is true if key was found, otherwise false.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree) Get(key []byte) (value []byte, found bool) {
	node, err := tree.lookup(key)
	if err != nil {
		return nil, false
	}
	if node != nil {
		return node.Item.Value, true
	}
	return nil, false
}

// Remove remove the node from the tree by key.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree) Remove(key []byte) {
	var child *Node
	node, err := tree.lookup(key)

	if err != nil || node == nil {
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
		// pred := left.maximumNode(tree)
		// node.Key = pred.Key
		// node.Item = pred.Item
		// node = pred
		node = left.maximumNode(tree)
	}

	if left == nil || right == nil {
		if right == nil {
			child = left
		} else {
			child = right
		}

		if node.Item.Color == black {
			node.Item.Color = nodeColor(child)
			tree.Save(node)

			tree.deleteCase1(node)
			// fmt.Println("update ", tree, node)
		}

		tree.replaceNode(node, child)

		if tree.emptyKey(node.ParentKey()) && child != nil {
			child.Item.Color = black
			tree.Save(child)
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
	current := tree.Root()
	for current != nil {
		parent = current
		current = current.Left(tree)
	}
	return parent
}

// Right returns the right-most (max) node or nil if tree is empty.
func (tree *Tree) Right() *Node {
	var parent *Node
	current := tree.Root()
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
	node := tree.Root()
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
	node := tree.Root()
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
	output(tree, tree.Root(), "", true, &str)
	// }
	return str
}

func output(tree *Tree, node *Node, prefix string, isTail bool, str *string) {
	// fmt.Printf("Node : %v+\n", node)
	if node == nil {
		return
	}
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

func (tree *Tree) lookup(key []byte) (*Node, error) {

	return tree.GetNode(key)
	// node := tree.Root()
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
	node.RightKey(right.LeftKey())
	if !tree.emptyKey(right.LeftKey()) {
		rightLeft := right.Left(tree)
		rightLeft.ParentKey(node.Key)
		tree.Save(rightLeft)
	}
	right.LeftKey(node.Key)
	node.ParentKey(right.Key)
	tree.Save(node)
	tree.Save(right)
}

func (tree *Tree) rotateRight(node *Node) {
	left := node.Left(tree)
	tree.replaceNode(node, left)
	node.LeftKey(left.RightKey())
	if !tree.emptyKey(left.RightKey()) {
		leftRight := left.Right(tree)
		leftRight.ParentKey(node.Key)
		tree.Save(leftRight)
	}
	left.RightKey(node.Key)
	node.ParentKey(left.Key)
	tree.Save(node)
	tree.Save(left)
}

func (tree *Tree) replaceNode(old *Node, new *Node) {

	newKey := emptyKey
	if new != nil {
		newKey = new.Key
	}

	if tree.emptyKey(old.ParentKey()) {
		// tree.Root = new
		tree.rootKey = newKey
	} else {
		// update left and right for oldParent
		oldParent := old.Parent(tree)
		// if new != nil {

		// fmt.Printf("Update new %v\n", new)
		// fmt.Printf("Update old parent %v\n", oldParent)
		if tree.Comparator(old.Key, oldParent.LeftKey()) == 0 {
			oldParent.LeftKey(newKey)
		} else {
			// remove oldParent right
			oldParent.RightKey(newKey)
		}
		// fmt.Printf("Update old parent %v\n", oldParent)
		tree.Save(oldParent)
		// }
		// fmt.Println("Replace tree node", old, new, oldParent)
	}
	if new != nil {
		// here is the swap, not update key
		// new.Parent = old.Parent
		new.ParentKey(old.ParentKey())
		tree.Save(new)
	}

	// fmt.Println("Final tree", tree)

}

func (tree *Tree) insertCase1(node *Node) {

	// fmt.Printf("Insert case1 :%s\n", node)
	if tree.emptyKey(node.ParentKey()) {
		node.Item.Color = black
		// store this
		// tree.Save(node)
		// fmt.Println("Breaking case1")
	} else {
		tree.insertCase2(node)
	}
}

func (tree *Tree) insertCase2(node *Node) {
	parent := node.Parent(tree)
	// fmt.Printf("Insert case 2, parent: %s", parent)
	if nodeColor(parent) == black {
		// tree.Save(node)
		// fmt.Println("Breaking case2")
		return
	}

	tree.insertCase3(node)
}

func (tree *Tree) insertCase3(node *Node) {
	parent := node.Parent(tree)
	uncle := node.uncle(tree)
	grandparent := node.grandparent(tree)
	// fmt.Printf("Insert case 3, uncle: %s\n", uncle)
	if nodeColor(uncle) == red {
		parent.Item.Color = black
		uncle.Item.Color = black
		tree.Save(uncle)
		// tree.Save(parent)
		grandparent.Item.Color = red
		tree.insertCase1(grandparent)
		tree.Save(grandparent)
	} else {
		tree.insertCase4(node)
	}
}

func (tree *Tree) insertCase4(node *Node) {
	grandparent := node.grandparent(tree)
	parent := node.Parent(tree)
	if tree.Comparator(node.Key, parent.RightKey()) == 0 &&
		tree.Comparator(parent.Key, grandparent.LeftKey()) == 0 {
		tree.rotateLeft(parent)
		node = node.Left(tree)
	} else if tree.Comparator(node.Key, parent.LeftKey()) == 0 &&
		tree.Comparator(parent.Key, grandparent.RightKey()) == 0 {
		tree.rotateRight(parent)
		node = node.Right(tree)
	}

	tree.insertCase5(node)
}

func (tree *Tree) insertCase5(node *Node) {
	grandparent := node.grandparent(tree)
	parent := node.Parent(tree)
	parent.Item.Color = black
	grandparent.Item.Color = red
	tree.Save(parent)
	tree.Save(grandparent)
	// fmt.Printf("insertCase5 :%s | %s | %s | %s \n", node.Key, parent.LeftKey(), parent, grandparent.LeftKey())
	// fmt.Printf("insertCase5 :%s | %s \n", parent.RightKey(), grandparent.Right(tree))

	if tree.Comparator(node.Key, parent.LeftKey()) == 0 &&
		tree.Comparator(parent.Key, grandparent.LeftKey()) == 0 {
		tree.rotateRight(grandparent)
	} else if tree.Comparator(node.Key, parent.RightKey()) == 0 &&
		tree.Comparator(parent.Key, grandparent.RightKey()) == 0 {
		tree.rotateLeft(grandparent)
	}

}

func (tree *Tree) Save(node *Node) error {
	// value, err := json.Marshal(node.Item)
	value, err := tree.EncodeToBytes(node.Item)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Printf("Save %s, value :%x\n", node.Key, value)
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
	parent := node.Parent(tree)
	sibling := node.sibling(tree)

	if nodeColor(sibling) == red {
		parent.Item.Color = red
		sibling.Item.Color = black
		tree.Save(parent)
		tree.Save(sibling)
		if tree.Comparator(node.Key, parent.LeftKey()) == 0 {
			tree.rotateLeft(parent)
		} else {
			tree.rotateRight(parent)
		}
	}

	tree.deleteCase3(node)
}

func (tree *Tree) deleteCase3(node *Node) {

	parent := node.Parent(tree)
	sibling := node.sibling(tree)
	siblingLeft := sibling.Left(tree)
	siblingRight := sibling.Right(tree)

	if nodeColor(parent) == black &&
		nodeColor(sibling) == black &&
		nodeColor(siblingLeft) == black &&
		nodeColor(siblingRight) == black {
		sibling.Item.Color = red
		tree.Save(sibling)
		tree.deleteCase1(parent)

		fmt.Println("delete node", string(node.Key), parent)

		tree.deleteNode(node, false)

	} else {
		tree.deleteCase4(node)
	}

}

func (tree *Tree) deleteCase4(node *Node) {
	parent := node.Parent(tree)
	sibling := node.sibling(tree)
	siblingLeft := sibling.Left(tree)
	siblingRight := sibling.Right(tree)

	if nodeColor(parent) == red &&
		nodeColor(sibling) == black &&
		nodeColor(siblingLeft) == black &&
		nodeColor(siblingRight) == black {
		sibling.Item.Color = red
		parent.Item.Color = black
		tree.Save(sibling)
		tree.Save(parent)
	} else {
		tree.deleteCase5(node)
	}
}

func (tree *Tree) deleteCase5(node *Node) {
	parent := node.Parent(tree)
	sibling := node.sibling(tree)
	siblingLeft := sibling.Left(tree)
	siblingRight := sibling.Right(tree)

	if tree.Comparator(node.Key, parent.LeftKey()) == 0 &&
		nodeColor(sibling) == black &&
		nodeColor(siblingLeft) == red &&
		nodeColor(siblingRight) == black {
		sibling.Item.Color = red
		siblingLeft.Item.Color = black

		tree.Save(sibling)
		tree.Save(siblingLeft)

		tree.rotateRight(sibling)

	} else if tree.Comparator(node.Key, parent.RightKey()) == 0 &&
		nodeColor(sibling) == black &&
		nodeColor(siblingRight) == red &&
		nodeColor(siblingLeft) == black {
		sibling.Item.Color = red
		siblingRight.Item.Color = black

		tree.Save(sibling)
		tree.Save(siblingRight)

		tree.rotateLeft(sibling)

	}

	tree.deleteCase6(node)
}

func (tree *Tree) deleteCase6(node *Node) {
	parent := node.Parent(tree)
	sibling := node.sibling(tree)
	siblingLeft := sibling.Left(tree)
	siblingRight := sibling.Right(tree)

	sibling.Item.Color = nodeColor(parent)
	parent.Item.Color = black

	tree.Save(sibling)
	tree.Save(parent)

	fmt.Println("before-update ", tree, sibling, parent, siblingLeft, siblingRight)

	if tree.Comparator(node.Key, parent.LeftKey()) == 0 && nodeColor(siblingRight) == red {
		siblingRight.Item.Color = black
		tree.Save(siblingRight)
		tree.rotateLeft(parent)
		// parent.LeftKey(emptyKey)
	} else if nodeColor(siblingLeft) == red {
		siblingLeft.Item.Color = black
		tree.Save(siblingLeft)
		tree.rotateRight(parent)
		// parent.RightKey(emptyKey)
	}

	// update the parent meta then delete the current node from db
	// tree.Save(parent)
	tree.deleteNode(node, false)
	fmt.Println("update ", tree, parent, sibling)
}

func nodeColor(node *Node) bool {
	if node == nil {
		return black
	}
	return node.Item.Color
}

func (tree *Tree) deleteNode(node *Node, force bool) {
	// update parent
	// parent := node.Parent(tree)
	// fmt.Println("Update parent", parent, "Delete node", node)
	// if tree.Comparator(node.Key, parent.LeftKey()) == 0 {
	// 	parent.LeftKey(emptyKey)
	// } else {
	// 	parent.RightKey(emptyKey)
	// }
	// tree.Save(parent)

	if force {
		tree.db.Delete(node.Key)
	} else {
		node.Item.Deleted = true
		tree.Save(node)
	}
}
