package redblacktree

import (
	"fmt"
)

type KeyMeta struct {
	Left   []byte
	Right  []byte
	Parent []byte
}

func (keys *KeyMeta) String() string {
	return fmt.Sprintf("L: %v, P: %v, R: %v", string(keys.Left), string(keys.Parent), string(keys.Right))
}

type Item struct {
	Keys  *KeyMeta
	Value []byte
	Color color
}

// Node is a single element within the tree
type Node struct {
	Key  []byte
	Item *Item
}

func (node *Node) String() string {
	return fmt.Sprintf("%v -> %v, (%v)\n", string(node.Key), string(node.Value()), node.Item.Keys.String())
}

func (node *Node) maximumNode(tree *Tree) *Node {
	if node == nil {
		return nil
	}
	for node.Item.Keys.Right != nil {
		node = node.Right(tree)
	}
	return node
}

func (node *Node) LeftKey(keys ...[]byte) []byte {
	if node.Item == nil || node.Item.Keys == nil {
		return nil
	}
	if len(keys) == 1 {
		node.Item.Keys.Left = keys[0]
	}

	return node.Item.Keys.Left
}

func (node *Node) RightKey(keys ...[]byte) []byte {
	if node.Item == nil || node.Item.Keys == nil {
		return nil
	}
	if len(keys) == 1 {
		if string(node.Item.Keys.Right) == "3" {
			fmt.Printf("RightKey :%s\n", keys[0])
		}
		node.Item.Keys.Right = keys[0]
	}

	return node.Item.Keys.Right
}

func (node *Node) ParentKey(keys ...[]byte) []byte {
	if node.Item == nil || node.Item.Keys == nil {
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
	if node != nil && node.Item.Keys.Parent != nil {
		return node.Parent(tree).Parent(tree)
	}
	return nil
}

func (node *Node) uncle(tree *Tree) *Node {
	if node == nil || node.Item.Keys.Parent == nil || node.Parent(tree).Item.Keys.Parent == nil {
		return nil
	}
	return node.Parent(tree).sibling(tree)
}

func (node *Node) sibling(tree *Tree) *Node {
	if node == nil || node.Item.Keys.Parent == nil {
		return nil
	}
	parent := node.Parent(tree)

	if tree.Comparator(node.Key, parent.Item.Keys.Left) == 0 {
		return parent.Right(tree)
	}
	return parent.Left(tree)
}
