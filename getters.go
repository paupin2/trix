package trix

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	errorNodeNotFound = fmt.Errorf("node not found")
)

// GetNodes returns a slice with the nodes that match the spec.
func (node *Node) GetNodes(keys ...interface{}) NodeList {
	return internalGetNodes(node, ParseKeys(keys), 0)
}

// ERROR GETTERS
// These return node values, converted do different data types for convenience.
// If no matching node is found return `errorNodeNotFound`.
// If there is a conversion error, return it.

// TryGet returns value for the first node matching the spec; if it can't find
// any, an error is returned.
func (node *Node) TryGet(keys ...interface{}) (Value, error) {
	childNode, err := node.TryGetNode(keys...)
	if err == nil {
		return childNode.Value, nil
	}
	return nil, err
}

// TryGetNode returns the first node matching the spec; if it can't find any,
// an error is returned.
func (node *Node) TryGetNode(keys ...interface{}) (*Node, error) {
	return internalTryGetNode(node, ParseKeys(keys))
}

// TryGetString returns value for the first node matching the spec, converted to
// a string; if it can't find a value or if here's a conversion error,
// an error is returned instead.
func (node *Node) TryGetString(keys ...interface{}) (string, error) {
	childNode, err := node.TryGetNode(keys...)
	if err != nil {
		return "", err
	}
	return childNode.internalStringValue(), nil
}

// TryGetInt returns value for the first node matching the spec, converted to
// an int; if it can't find a value or if here's a conversion error,
// an error is returned instead.
func (node *Node) TryGetInt(keys ...interface{}) (int, error) {
	if v, err := node.TryGet(keys...); err != nil {
		return 0, err
	} else if castd, ok := v.(int); ok {
		return castd, nil
	} else {
		return parseInt(v)
	}
}

// TryGetFloat returns value for the first node matching the spec, converted to
// an int; if it can't find a value or if here's a conversion error,
// an error is returned instead.
func (node *Node) TryGetFloat(keys ...interface{}) (float64, error) {
	if v, err := node.TryGet(keys...); err != nil {
		return 0, err
	} else if castd, ok := v.(float64); ok {
		return castd, nil
	} else {
		return strconv.ParseFloat(fmt.Sprint(v), 64)
	}
}

// TryGetBool returns value for the first node matching the spec, converted to
// a bool; if it can't find a value or if here's a conversion error,
// an error is returned instead.
func (node *Node) TryGetBool(keys ...interface{}) (bool, error) {
	if v, err := node.TryGet(keys...); err != nil {
		return false, err
	} else if castd, ok := v.(bool); ok {
		return castd, nil
	} else {
		return parseBool(v)
	}
}

// TryGetDuration returns value for the first node matching the spec, converted to
// a duraion; if it can't find a value or if here's a conversion error,
// an error is returned instead.
func (node *Node) TryGetDuration(keys ...interface{}) (time.Duration, error) {
	if v, err := node.TryGet(keys...); err != nil {
		return 0, err
	} else if castd, ok := v.(time.Duration); ok {
		return castd, nil
	} else {
		return parseDuration(v)
	}
}

// TryGetTime returns value for the first node matching the spec, converted to
// a duraion; if it can't find a value or if here's a conversion error,
// an error is returned instead.
func (node *Node) TryGetTime(keys ...interface{}) (time.Time, error) {
	if v, err := node.TryGet(keys...); err != nil {
		return time.Time{}, err
	} else if castd, ok := v.(time.Time); ok {
		return castd, nil
	} else {
		return parseTime(v)
	}
}

// DEFAULT GETTERS
// These return node values, converted do different data types for convenience;
// in case of 0 results or conversion errors, return the default value.

// GetNodeDefault returns the first node that matches the spec.
// If no node matches, return the default value instead.
func (node *Node) GetNodeDefault(def *Node, keys ...interface{}) *Node {
	if val, err := node.TryGetNode(keys...); err == nil {
		return val
	}
	return def
}

// GetDefault returns the value of the first node that matches the spec.
// If no node matches, return the default value instead.
func (node *Node) GetDefault(def Value, keys ...interface{}) Value {
	if val, err := node.TryGet(keys...); err == nil {
		return val
	}
	return def
}

// GetStringDefault returns the value of the first node that matches the spec.
// If no node matches, return the default value instead.
func (node *Node) GetStringDefault(def string, keys ...interface{}) string {
	if val, err := node.TryGetString(keys...); err == nil {
		return val
	}
	return def
}

// GetIntDefault returns the value of the first node that matches the spec,
// converted to an int. If no node matches, or converting fails, return
// the default value instead.
func (node *Node) GetIntDefault(def int, keys ...interface{}) int {
	if val, err := node.TryGetInt(keys...); err == nil {
		return val
	}
	return def
}

// GetFloatDefault returns the value of the first node that matches the spec,
// converted to a float64. If no node matches, or converting fails, return
// the default value instead.
func (node *Node) GetFloatDefault(def float64, keys ...interface{}) float64 {
	if val, err := node.TryGetFloat(keys...); err == nil {
		return val
	}
	return def
}

// GetBoolDefault returns the value of the first node that matches the spec,
// converted to a bool. If no node matches, or converting fails, return
// the default value instead.
func (node *Node) GetBoolDefault(def bool, keys ...interface{}) bool {
	if val, err := node.TryGetBool(keys...); err == nil {
		return val
	}
	return def
}

// GetDurationDefault returns the value of the first node that matches the spec,
// converted to a duration. If no node matches, or converting fails, return
// the default value instead.
func (node *Node) GetDurationDefault(def time.Duration, keys ...interface{}) time.Duration {
	if val, err := node.TryGetDuration(keys...); err == nil {
		return val
	}
	return def
}

// SIMPLE GETTERS
// These return node values, converted do different data types for convenience;
// in case of 0 results or conversion errors, return the type's default value.

// GetNode returns the first node that matches the spec.
// If no node matches, return nil.
func (node *Node) GetNode(keys ...interface{}) *Node {
	return node.GetNodeDefault(nil, keys...)
}

// Get returns the value of the first node that matches the spec.
// If no node matches, return the type's default value instead.
// If no argument is given, the current node's value is returned.
func (node *Node) Get(keys ...interface{}) Value {
	val, _ := node.TryGet(keys...)
	return val
}

// GetString returns the value of the first node that matches the spec,
// converted to a string. If no node matches, or converting fails, return
// the type's default value instead.
// If no argument is given, the current node is used.
func (node *Node) GetString(keys ...interface{}) string {
	val, _ := node.TryGetString(keys...)
	return val
}

// GetInt returns the value of the first node that matches the spec,
// converted to an int. If no node matches, or converting fails, return
// the type's default value instead.
// If no argument is given, the current node is used.
func (node *Node) GetInt(keys ...interface{}) int {
	val, _ := node.TryGetInt(keys...)
	return val
}

// GetFloat returns the value of the first node that matches the spec,
// converted to an int. If no node matches, or converting fails, return
// the type's default value instead.
// If no argument is given, the current node is used.
func (node *Node) GetFloat(keys ...interface{}) float64 {
	val, _ := node.TryGetFloat(keys...)
	return val
}

// GetBool returns the value of the first node that matches the spec,
// converted to a bool. If no node matches, or converting fails, return
// the type's default value instead.
// If no argument is given, the current node is used.
func (node *Node) GetBool(keys ...interface{}) bool {
	val, _ := node.TryGetBool(keys...)
	return val
}

// GetDuration returns the value of the first node that matches the spec,
// converted to a duration. If no node matches, or converting fails, return
// the type's default value instead.
// If no argument is given, the current node is used.
func (node *Node) GetDuration(keys ...interface{}) time.Duration {
	val, _ := node.TryGetDuration(keys...)
	return val
}

// GetTime returns the value of the first node that matches the spec,
// converted to a timestamp. If no node matches, or converting fails, return
// the type's default value instead.
// If no argument is given, the current node is used.
func (node *Node) GetTime(keys ...interface{}) time.Time {
	val, _ := node.TryGetTime(keys...)
	return val
}

// MUST GETTERS
// These return node values, converted do different data types for convenience;
// in case of 0 results or conversion errors, panic. These should not be

// MustGetNode returns the first node that matches the spec. If no node
// matches, panic. This is most suited for intializations.
func (node *Node) MustGetNode(keys ...interface{}) *Node {
	val, err := node.TryGetNode(keys...)
	if err != nil {
		panic(fmt.Sprintf("Required conf key %s: %v",
			strings.Join(ParseKeys(keys), "."),
			err,
		))
	}
	return val
}

// MustGet returns the value of the first node that matches the spec. If no node
// matches, panic. This is most suited for intializations.
func (node *Node) MustGet(keys ...interface{}) Value {
	val, err := node.TryGet(keys...)
	if err != nil {
		panic(fmt.Sprintf("Required conf key %s: %v",
			strings.Join(ParseKeys(keys), "."),
			err,
		))
	}
	return val
}

// MustGetString returns the value of the first node that matches the spec,
// converted to a string. If no node matches, or converting fails, panic.
// This is most suited for intializations.
func (node *Node) MustGetString(keys ...interface{}) string {
	val, err := node.TryGetString(keys...)
	if err != nil {
		panic(fmt.Sprintf("Required conf key %s: %v",
			strings.Join(ParseKeys(keys), "."),
			err,
		))
	}
	return val
}

// MustGetInt returns the value of the first node that matches the spec,
// converted to an int. If no node matches, or converting fails, panic.
// This is most suited for intializations.
func (node *Node) MustGetInt(keys ...interface{}) int {
	val, err := node.TryGetInt(keys...)
	if err != nil {
		panic(fmt.Sprintf("Required conf key %s: %v",
			strings.Join(ParseKeys(keys), "."),
			err,
		))
	}
	return val
}

// MustGetFloat returns the value of the first node that matches the spec,
// converted to an float64. If no node matches, or converting fails, panic.
// This is most suited for intializations.
func (node *Node) MustGetFloat(keys ...interface{}) float64 {
	val, err := node.TryGetFloat(keys...)
	if err != nil {
		panic(fmt.Sprintf("Required conf key %s: %v",
			strings.Join(ParseKeys(keys), "."),
			err,
		))
	}
	return val
}

// MustGetBool returns the value of the first node that matches the spec,
// converted to a bool. If no node matches, or converting fails, panic.
// This is most suited for intializations.
func (node *Node) MustGetBool(keys ...interface{}) bool {
	val, err := node.TryGetBool(keys...)
	if err != nil {
		panic(fmt.Sprintf("Required conf key %s: %v",
			strings.Join(ParseKeys(keys), "."),
			err,
		))
	}
	return val
}

// MustGetDuration returns the value of the first node that matches the spec,
// converted to a duration. If no node matches, or converting fails, panic.
// This is most suited for intializations.
func (node *Node) MustGetDuration(keys ...interface{}) time.Duration {
	val, err := node.TryGetDuration(keys...)
	if err != nil {
		panic(fmt.Sprintf("Required conf key %s: %v",
			strings.Join(ParseKeys(keys), "."),
			err,
		))
	}
	return val
}

// EXTRA GETTERS

// GetValues return the values of all of the nodes that match the spec.
func (node *Node) GetValues(keys ...interface{}) []Value {
	values := make([]Value, 0, 10)
	for _, node := range node.GetNodes(keys...) {
		if node.IsLeaf() {
			values = append(values, node.Value)
		}
	}
	return values
}

// GetMap returns a key/value pair for a spec like "*.*.common.region.*.name".
// Use the position of the last star as the key, and the node's value.
func (node *Node) GetMap(keys ...interface{}) Args {
	if len(keys) == 0 {
		return node.GetMap("*")
	}

	// split the original spec in two, one before and one after the last `*`
	lastStarPos := 0
	parseKeys := ParseKeys(keys)
	ifParsedKeys := make([]interface{}, len(parseKeys))
	for index, part := range parseKeys {
		ifParsedKeys[index] = parseKeys[index]
		if part == "*" {
			lastStarPos = index
		}
	}
	keysUntilStar := ifParsedKeys[:lastStarPos+1]
	keysAfterStar := ifParsedKeys[lastStarPos+1:]

	// build the result map
	result := Args{}
	for _, subnode := range node.GetNodes(keysUntilStar...) {
		key := subnode.Key
		if len(keysAfterStar) > 0 {
			subnode = subnode.GetNode(keysAfterStar...)
		}
		if subnode == nil {
			continue
		}
		result[key] = subnode.internalStringValue()
	}
	return result
}

// GetStringMap returns a map for a spec like "*.*.common.region.*.name".
// Use the position of the last star as the key, and the node's string value.
func (node *Node) GetStringMap(keys ...interface{}) StrArgs {
	result := map[string]string{}
	for key, value := range node.GetMap(keys...) {
		result[key] = fmt.Sprint(value)
	}
	return result
}

// GetStringValues returns a slice with values for all matching node values.
func (node *Node) GetStringValues(keys ...interface{}) []string {
	found := node.GetNodes(keys...)
	result := make([]string, len(found))
	for i, subnode := range found {
		result[i] = subnode.internalStringValue()
	}
	return result
}
