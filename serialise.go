package trix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// MarshalJSON returns the node node's and its descendants' representation
// in JSON.
func (node *Node) MarshalJSON() ([]byte, error) {
	if node == nil {
		return []byte{}, nil
	}

	forceArray := node.Flags&ForceArray > 0
	forceMap := node.Flags&ForceMap > 0
	if len(node.Children) == 0 && !forceArray && !forceMap {
		return json.Marshal(node.Value)
	}

	if forceArray || (!forceMap && node.hasOnlyNumericKeys()) {
		// return a sorted array
		children := make([]interface{}, len(node.ChildKeys))
		for index, key := range node.ChildKeys {
			children[index] = node.Children[key]
		}
		return json.Marshal(children)
	}

	// serialise children as a sorted map
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	buf.Write([]byte{'{'})
	for i, key := range node.ChildKeys {
		if i > 0 {
			buf.WriteByte(',')
		}
		enc.Encode(key)
		buf.Write([]byte{':'})
		enc.Encode(node.Children[key])
	}
	buf.Write([]byte{'}', '\n'})
	return buf.Bytes(), nil
}

// Dump dumps the JSON representation of a node and its descendants.
func (node *Node) Dump(w io.Writer, short bool) {
	if node == nil {
		return
	}

	formatValue := func(v Value) string {
		if s, ok := v.(string); ok {
			return s
		} else if t, ok := v.(time.Time); ok {
			return t.Format(time.RFC3339Nano)
		}
		return fmt.Sprint(v)
	}

	var toString func(*Node, int)
	toString = func(node *Node, depth int) {
		if short && depth > 0 {
			fmt.Fprintf(w, "%s=", node.Key)
		}
		if short && node.Value != nil && depth > 0 {
			w.Write([]byte(formatValue(node.Value)))
		}
		if len(node.ChildKeys) > 0 {
			if short && depth > 0 {
				w.Write([]byte("{"))
			}
			for i, k := range node.ChildKeys {
				if short && i > 0 {
					w.Write([]byte(","))
				}
				toString(node.Children[k], depth+1)
			}
			if short && depth > 0 {
				w.Write([]byte("}"))
			}
		} else if !short {
			fmt.Fprintf(w, "%s=%s\n", strings.Join(node.Path(), "."), formatValue(node.Value))
		}
	}

	if short {
		w.Write([]byte("{"))
	}
	toString(node, 0)
	if short {
		w.Write([]byte("}"))
	}
}
