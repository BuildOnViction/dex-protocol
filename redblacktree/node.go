// Copyright (c) 2019, Agiletech Viet Nam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redblacktree

import (
	"fmt"
)

const (
	black, red bool = true, false
)

type KeyMeta struct {
	Left   []byte
	Right  []byte
	Parent []byte
}

func formatBytes(key []byte) string {
	if len(key) == 0 || key == nil {
		return "<nil>"
	}
	return string(key)
}

func (keys *KeyMeta) String() string {
	return fmt.Sprintf("L: %v, P: %v, R: %v", formatBytes(keys.Left), formatBytes(keys.Parent), formatBytes(keys.Right))
}

type Item struct {
	Keys    *KeyMeta
	Value   []byte
	Deleted bool
	Color   bool
}

// Node is a single element within the tree
type Node struct {
	Key  []byte
	Item *Item
}

func (node *Node) String() string {
	if node == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v -> %v, (%v)\n", string(node.Key), string(node.Value()), node.Item.Keys.String())
}

func (node *Node) maximumNode(tree *Tree) *Node {
	if node == nil {
		return nil
	}
	for !tree.emptyKey(node.RightKey()) {
		node = node.Right(tree)
	}
	return node
}

func (node *Node) LeftKey(keys ...[]byte) []byte {
	if node == nil || node.Item == nil || node.Item.Keys == nil {
		return nil
	}
	if len(keys) == 1 {
		node.Item.Keys.Left = keys[0]
	}

	return node.Item.Keys.Left
}

func (node *Node) RightKey(keys ...[]byte) []byte {
	if node == nil || node.Item == nil || node.Item.Keys == nil {
		return nil
	}
	if len(keys) == 1 {
		// if string(node.Key) == "1" {
		// 	fmt.Printf("Update right key: %s\n", string(keys[0]))
		// if string(keys[0]) == "3" {
		// 	panic("should stops")
		// }
		// }
		node.Item.Keys.Right = keys[0]
	}

	return node.Item.Keys.Right
}

func (node *Node) ParentKey(keys ...[]byte) []byte {
	if node == nil || node.Item == nil || node.Item.Keys == nil {
		return nil
	}
	if len(keys) == 1 {
		node.Item.Keys.Parent = keys[0]
	}

	return node.Item.Keys.Parent
}

func (node *Node) Left(tree *Tree) *Node {
	key := node.LeftKey()

	node, err := tree.GetNode(key)
	if err != nil {
		fmt.Println(err)
	}

	return node
}

func (node *Node) Right(tree *Tree) *Node {
	key := node.RightKey()
	node, err := tree.GetNode(key)
	if err != nil {
		fmt.Println(err)
	}
	return node
}

func (node *Node) Parent(tree *Tree) *Node {
	key := node.ParentKey()
	node, err := tree.GetNode(key)
	if err != nil {
		fmt.Println(err)
	}
	return node
}

func (node *Node) Value() []byte {
	return node.Item.Value
}

func (node *Node) grandparent(tree *Tree) *Node {
	if node != nil && !tree.emptyKey(node.ParentKey()) {
		return node.Parent(tree).Parent(tree)
	}
	return nil
}

func (node *Node) uncle(tree *Tree) *Node {
	if node == nil || tree.emptyKey(node.ParentKey()) {
		return nil
	}
	parent := node.Parent(tree)
	// if tree.emptyKey(parent.ParentKey()) {
	// 	return nil
	// }

	return parent.sibling(tree)
}

func (node *Node) sibling(tree *Tree) *Node {
	if node == nil || tree.emptyKey(node.ParentKey()) {
		return nil
	}
	parent := node.Parent(tree)
	// fmt.Printf("Parent: %s\n", parent)
	if tree.Comparator(node.Key, parent.LeftKey()) == 0 {
		return parent.Right(tree)
	}
	return parent.Left(tree)
}
