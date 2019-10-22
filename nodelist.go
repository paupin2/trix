package trix

// NodeList represents a list of pointers to nodes
type NodeList []*Node

// ConvertValues applies the conversion function to each of the NodeList's
// nodes that match specified keys, and replaces its value with the one
// returned.
func (nodes NodeList) ConvertValues(conv func(*Node) Value, keys ...string) NodeList {
	if nodes == nil {
		return nodes
	}
	for _, node := range nodes {
		matches := len(keys) == 0
		for _, key := range keys {
			if node.Key == key {
				matches = true
				break
			}
		}
		if matches {
			node.Value = conv(node)
		}
	}
	return nodes
}

// ValuesToString converts values from the specified children of each node in the
// NodeList to string.
func (nodes NodeList) ValuesToString(keys ...string) NodeList {
	return nodes.ConvertValues(func(node *Node) Value {
		return node.GetString()
	}, keys...)
}

// ValuesToInt converts values from the specified children of each node in the
// NodeList to Int.
func (nodes NodeList) ValuesToInt(keys ...string) NodeList {
	return nodes.ConvertValues(func(node *Node) Value {
		return node.GetInt()
	}, keys...)

}

// ValuesToFloat converts values from the specified children of each node in the
// NodeList to Float.
func (nodes NodeList) ValuesToFloat(keys ...string) NodeList {
	return nodes.ConvertValues(func(node *Node) Value {
		return node.GetFloat()
	}, keys...)
}

// ValuesToBool converts values from the specified children of each node in the
// NodeList to Bool.
func (nodes NodeList) ValuesToBool(keys ...string) NodeList {
	return nodes.ConvertValues(func(node *Node) Value {
		return node.GetBool()
	}, keys...)
}

// ValuesToDuration converts values from the specified children of each node in the
// NodeList to Duration.
func (nodes NodeList) ValuesToDuration(keys ...string) NodeList {
	return nodes.ConvertValues(func(node *Node) Value {
		return node.GetDuration()
	}, keys...)
}

// ForEach runs the specified callback on each resulting node, and returns the
// resulting slice.
func (nodes NodeList) ForEach(cb func(node *Node) Value) []Value {
	result := make([]Value, len(nodes))
	for i := range nodes {
		result[i] = cb(nodes[i])
	}
	return result
}

// Filter runs the specified callback on each resulting node, and returns the
// nodes where the callback returns true.
func (nodes NodeList) Filter(cb func(node *Node) bool) NodeList {
	result := make(NodeList, 0, len(nodes))
	for _, node := range nodes {
		if cb(node) {
			result = append(result, node)
		}
	}
	return result
}

// FilterByValue returns the subset of the NodeList where the value equals
// the specified one.
func (nodes NodeList) FilterByValue(value Value) NodeList {
	return nodes.Filter(func(node *Node) bool {
		return node.Value == value
	})
}

// First returns the first node from the list, or nil if the list is empty.
func (nodes NodeList) First() *Node {
	if len(nodes) == 0 {
		return nil
	}
	return nodes[0]
}
