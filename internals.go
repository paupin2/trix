package trix

import (
	"fmt"
)

func (node *Node) internalStringValue() string {
	if node == nil || node.Value == nil {
		return ""
	} else if s, ok := node.Value.(string); ok {
		return s
	}
	return fmt.Sprint(node.Value)
}

func internalSet(node *Node, keys []string, value Value) *Node {
	if len(keys) == 0 {
		return nil
	}

	// find the node to update, creating intermediate nodes as necessary
	nodeToUpdate := node
	for _, key := range keys {
		child, found := nodeToUpdate.Children[key]
		if !found {
			child = NewNode(key)
			nodeToUpdate.Adopt(child)
		}

		// continue using this as the parent
		nodeToUpdate = child
	}

	// update the child's value
	if value != nil {
		nodeToUpdate.Value = value
	}
	return nodeToUpdate
}

// internalUnset will remove the specified node and return it
func internalUnset(node *Node, keys []string) *Node {
	if len(keys) > 0 {
		key, keys := keys[0], keys[1:]
		if child, found := node.Children[key]; found {
			if len(keys) > 0 {
				// this isn't the last key
				return internalUnset(child, keys)
			}

			// remove it from both lists
			delete(node.Children, key)
			for index, ck := range node.ChildKeys {
				if ck == key {
					node.ChildKeys = append(node.ChildKeys[:index], node.ChildKeys[index+1:]...)
					break
				}
			}
			child.Parent = nil
			return child
		}
	}
	return nil
}

// internalGetNodes will look for
func internalGetNodes(node *Node, parsedKeys []string, limit int) NodeList {
	result := NodeList{}
	if node == nil {
		// so that calling GetNodes from a nil node doesn't segfault
		return result
	} else if len(parsedKeys) == 0 {
		return NodeList{node}
	}

	var readNodes func(*Node, []string, int)
	readNodes = func(node *Node, spec []string, index int) {
		currentKey := spec[index]
		last := index+1 == len(spec)
		if currentKey == "*" {
			for _, key := range node.ChildKeys {
				childNode := node.Children[key]
				if last {
					result = append(result, childNode)
					if limit > 0 && len(result) >= limit {
						return
					}
				} else {
					readNodes(childNode, spec, index+1)
				}
			}
		} else {
			if childNode, found := node.Children[currentKey]; found {
				if last {
					result = append(result, childNode)
					if limit > 0 && len(result) >= limit {
						return
					}
				} else {
					readNodes(childNode, spec, index+1)
				}
			}
			// "*" works both ways; this handles "server.app" prefixes (usually *.*)
			if childNode, found := node.Children["*"]; found {
				if last {
					result = append(result, childNode)
					if limit > 0 && len(result) >= limit {
						return
					}
				} else {
					readNodes(childNode, spec, index+1)
				}
			}
		}
	}

	// if we have results from more than 1 scope, they will most likely not
	// be sorted; if this is an issue we can count the number of scopes with
	// results (when (count before `readNodes`) > count after) and if greater
	// than 1, sort `result`.
	for {
		readNodes(node, parsedKeys, 0)
		if limit > 0 && len(result) >= limit {
			break
		}

		// is there a parent scope where can also look?
		parentScope := node.GetRoot().Parent
		if parentScope == nil {
			break
		}

		if node.Flags&IsRoot == 0 {
			// the node is not a root, but a child; in order to try the parent
			// scope, we have to use the full/absolute path.
			nodePath := node.Path()
			absolutePath := make([]string, 0, len(nodePath)+len(parsedKeys))
			absolutePath = append(absolutePath, nodePath...)
			absolutePath = append(absolutePath, parsedKeys...)
			parsedKeys = absolutePath
		}

		// try again, using the parent scope as the new reference
		node = parentScope
	}

	return result
}

// internalTryGetNode will try o find the keys starting from the specified node.
func internalTryGetNode(node *Node, parsedKeys []string) (*Node, error) {
	if found := internalGetNodes(node, parsedKeys, 1); len(found) > 0 {
		return found[0], nil
	}
	return nil, errorNodeNotFound
}
