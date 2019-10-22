package trix

import (
	"encoding/json"
	"testing"
)

func TestMarshalJSON(t *testing.T) {
	root := NewRoot()
	root.SetKey("simple.int", 1)
	root.SetKey("simple.bool", true)
	root.SetKey("normal.array.1", "A")
	root.SetKey("normal.array.100", "B")
	root.SetKey("normal.array.020", "C")
	root.SetKey("normal.map.1", "apples")
	root.SetKey("normal.map.100", "oranges")
	root.SetKey("normal.map.twenty", "pears")
	root.SetKey("forced.map.1", "A")
	root.SetKey("forced.map.100", "B")
	root.SetKey("forced.map.020", "C")
	root.SetKey("forced.array.1", "apples")
	root.SetKey("forced.array.100", "oranges")
	root.SetKey("forced.array.twenty", "pears")

	check := func(expectedValue string) {
		t.Helper()
		byt, err := json.Marshal(root)
		testError(t, err, "")
		testEqualString(t, string(byt), expectedValue)
	}

	// preserve initial order
	check(`{"simple":{"int":1,"bool":true},"normal":{"array":["A","B","C"],"map":{"1":"apples","100":"oranges","twenty":"pears"}},"forced":{"map":["A","B","C"],"array":{"1":"apples","100":"oranges","twenty":"pears"}}}`)

	// sort
	root.SortRecursively()
	check(`{"forced":{"array":{"1":"apples","100":"oranges","twenty":"pears"},"map":["A","C","B"]},"normal":{"array":["A","C","B"],"map":{"1":"apples","100":"oranges","twenty":"pears"}},"simple":{"bool":true,"int":1}}`)

	// apply flags
	root.GetNode("forced.map").Flags = ForceMap
	root.GetNode("forced.array").Flags = ForceArray
	check(`{"forced":{"array":["apples","oranges","pears"],"map":{"1":"A","020":"C","100":"B"}},"normal":{"array":["A","C","B"],"map":{"1":"apples","100":"oranges","twenty":"pears"}},"simple":{"bool":true,"int":1}}`)

	root = NewRoot()
	root.AddNode("empty.array").Flags = ForceArray
	root.AddNode("empty.map").Flags = ForceMap
	check(`{"empty":{"array":[],"map":{}}}`)
}
