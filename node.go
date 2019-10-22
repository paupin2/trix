package trix

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
)

// NodeFlag is the type used to associate flags with a node
type NodeFlag byte

const (
	// NoFlags means the node has no flag
	NoFlags NodeFlag = 0

	// ForceMap means the node's direct children will be serialised as a map,
	// even if all keys are numeric
	ForceMap NodeFlag = 1 << (iota - 1)

	// ForceArray means the node's direct children will be serialised as
	// an array, even if some of the keys are not numeric
	ForceArray

	// IsRoot means the node is considered a Root node.
	// That is, `Parent` points to a parent tree, not a parent node.
	IsRoot
)

// Value is the type for a trix node
type Value interface{}

// Node represents a node
type Node struct {
	Key       string
	Value     Value
	Children  map[string]*Node
	ChildKeys []string
	Parent    *Node
	Flags     NodeFlag
}

// NewNode returns the pointer to a new, empty node.
func NewNode(key string) *Node {
	return &Node{
		Key:       key,
		Children:  map[string]*Node{},
		ChildKeys: []string{},
	}
}

// NewRoot returns a new, empty root node.
func NewRoot() *Node {
	root := NewNode("")
	root.Flags = IsRoot
	return root
}

// MustLoad is a convenient method to load
func MustLoad(filename string) *Node {
	root := NewRoot()
	if err := root.MergeFile(filename); err != nil {
		panic(fmt.Errorf("Could not load configuration from %s: %v", filename, err))
	}
	return root
}

// GetRoot returns the root for this node.
func (node *Node) GetRoot() *Node {
	p := node
	for ; p != nil && p.Parent != nil && p.Flags&IsRoot == 0; p = p.Parent {
	}
	return p
}

// Depth returns the depth of the node, that is, the number of parents it has.
// The minimum (root node) depth is 0.
func (node *Node) Depth() int {
	depth := 0
	for n := node; n != nil && n.Parent != nil && n.Flags&IsRoot == 0; n = n.Parent {
		depth++
	}
	return depth
}

// Path returns the path up to (and including) this node, as a string slice.
func (node *Node) Path() []string {
	depth := node.Depth()
	path := make([]string, depth)
	for n := node; n != nil; n = n.Parent {
		depth--
		if n.Key != "" {
			path[depth] = n.Key
		}
	}
	return path
}

// With returns a new child root tree with the specified arguments,
// that also inherits all values from the original one.
func (node *Node) With(args ...Args) *Node {
	root := node.GetRoot()
	newRoot := NewRoot()
	newRoot.Parent = root

	// if this is not called from the root, a new node should be created
	// to contain the arguments
	argsTarget := newRoot
	if root != node {
		argsTarget = internalSet(newRoot, node.Path(), nil)
	}

	// add all arguments
	for _, arg := range args {
		for key, value := range arg {
			argsTarget.SetKey(key, value)
		}
	}

	return newRoot
}

// FromArgs returns a new root node from an args structure.
func FromArgs(args Args) *Node {
	root := NewRoot()
	for key, value := range args {
		root.SetKey(key, value)
	}
	return root
}

// Rename changes the node's key. It does ensure the parent node is kept sorted.
func (node *Node) Rename(newKey string) *Node {
	if node != nil {
		if parent := node.Parent; parent != nil {
			parent.Unset(node.Key)
			node.Key = newKey
			parent.Adopt(node)
		}
	}
	return node
}

// IsLeaf returns whether the node is a left one (has no children).
func (node *Node) IsLeaf() bool {
	return len(node.ChildKeys) == 0
}

// Adopt the new child into the node's children, removing it from the previous
// parent if necessary.
func (node *Node) Adopt(child *Node) {
	// sever link with former parent
	if p := child.Parent; p != nil {
		p.Unset(child.Key)
	}

	if other, found := node.Children[child.Key]; found {
		// there's another child with the same key; remove it
		node.Unset(other.Key)
	}

	// add the child, update its parent and depth
	node.Children[child.Key] = child
	node.ChildKeys = append(node.ChildKeys, child.Key)
	child.Parent = node
}

// Merge a new subnode into the current one. Recursively create clones of each
// node as necessary. Any existing nodes that aren't overwritten are kept.
// Return the either newly-created or existing node.
func (node *Node) Merge(original *Node) *Node {
	if original == nil {
		return nil
	}

	// ensure the node exists
	old := node.GetNode(original.Key)
	if old == nil {
		old = NewNode(original.Key)
		old.Parent = node
		node.Adopt(old)
		node.Sort()
	}

	// overwrite the value
	old.Value = original.Value

	// merge children
	for _, key := range original.ChildKeys {
		old.Merge(original.Children[key])
	}

	return old
}

// hasOnlyNumericKeys returns whether the node only has numeric keys
func (node *Node) hasOnlyNumericKeys() bool {
	for _, key := range node.ChildKeys {
		if _, err := strconv.Atoi(key); err != nil {
			return false
		}
	}
	return true
}

// Sort sorts a node's children by their keys.
// Nodes with only integer keys are sorted numerically,
// while others are sorted alphabetically.
func (node *Node) Sort() {
	if node.hasOnlyNumericKeys() {
		NumericStringSlice(node.ChildKeys).Sort()
	} else {
		sort.StringSlice(node.ChildKeys).Sort()
	}
}

// SortRecursively will recursively sorts a node's children by their keys.
// Nodes with only integer keys are sorted numerically,
// while others are sorted alphabetically.
func (node *Node) SortRecursively() {
	node.Sort()
	for _, child := range node.Children {
		if len(child.Children) > 0 {
			child.SortRecursively()
		}
	}
}

// String returns the string representation of a node and its descendants.
func (node *Node) String() string {
	if node == nil {
		return ""
	}
	var buffer bytes.Buffer
	node.Dump(&buffer, true)
	return buffer.String()
}

// Set a child node with the specified value.
func (node *Node) Set(keys []interface{}, value Value) *Node {
	return internalSet(node, ParseKeys(keys), value)
}

// SetKey sets a child node with the specified value.
func (node *Node) SetKey(key string, value Value) *Node {
	return internalSet(node, ParseKeys([]interface{}{key}), value)
}

// FillKey will, on the first call, set the node's value. On subsequent calls
// it will convert the node from a list to a node, and add additional items.
// more than one value
func (node *Node) FillKey(keys string, value Value) *Node {
	childNode := internalSet(node, ParseKeys([]interface{}{keys}), nil) // get/create the child node
	var newNode *Node
	if len(childNode.ChildKeys) == 0 {
		if childNode.Value == nil {
			// the node has just been created; set its value
			newNode = childNode
		} else {
			// node has a value; convert original value to a child, and push the second one
			childNode.Push().Value = childNode.Value
			childNode.Value = nil
		}
	}
	if newNode == nil {
		newNode = childNode.Push()
	}
	newNode.Value = value
	return newNode
}

// AddNode adds a child node.
func (node *Node) AddNode(keys ...interface{}) *Node {
	return node.Set(keys, nil)
}

// Push adds a new child node, usinf a unique number as the next ID.
// This is usefull for fillin-in arrays.
// Return the newly-created node.
func (node *Node) Push() *Node {
	id := len(node.ChildKeys)
	for {
		id++
		sid := fmt.Sprint(id)
		if _, found := node.Children[sid]; found {
			// index already used
			continue
		}
		return node.SetKey(sid, nil)
	}
}

// PushValues adds all specified values as subnodes, using unique number as IDs.
// Return the original node.
func (node *Node) PushValues(values ...Value) *Node {
	for _, value := range values {
		node.Push().Value = value
	}
	return node
}

// Unset the child with the specified key, and return it.
// If the child is not found, return nil.
func (node *Node) Unset(keys ...interface{}) *Node {
	return internalUnset(node, ParseKeys(keys))
}
