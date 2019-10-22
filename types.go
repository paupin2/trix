package trix

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
)

// StrArgs is a string-string map
type StrArgs map[string]string

// NumericStringSlice represents a string slice that can be sorted using the
// integer representation of its values
type NumericStringSlice []string

// Sort this slice.
func (s NumericStringSlice) Sort()         { sort.Sort(s) }
func (s NumericStringSlice) Len() int      { return len(s) }
func (s NumericStringSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s NumericStringSlice) Less(i, j int) bool {
	ii, _ := strconv.Atoi(s[i])
	ij, _ := strconv.Atoi(s[j])
	return ii < ij
}

// Args represents a generic string-interface{} map
type Args map[string]interface{}

// Merge the map with another, adding or overwriting keys
func (args Args) Merge(other Args) Args {
	for key, value := range other {
		args[key] = value
	}
	return args
}

// Clone returns a clone of the original one
func (args Args) Clone() Args {
	n := Args{}
	for k, v := range args {
		n[k] = v
	}
	return n
}

// Add returns a new map, adding or overwriting keys
func (args Args) Add(other Args) Args {
	return args.Clone().Merge(other)
}

// GetString returns the specified key as a string
func (args Args) GetString(key string) string {
	if v, found := args[key]; found {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprint(v)
	}
	return ""
}

// String returns a simple string representation of the arguments, with the
// keys sorted. This is mainly convenient for testing.
func (args Args) String() string {
	// get sorted list of keys
	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}
	sort.StringSlice(keys).Sort()

	// write keys/values in a format similar to `fmt.printValue`
	buf := bytes.Buffer{}
	buf.WriteString("args[")
	first := true
	for _, key := range keys {
		if !first {
			buf.WriteString(" ")
		}
		first = false
		fmt.Fprintf(&buf, `%s:%v`, key, args[key])
	}
	buf.WriteString("]")
	return buf.String()
}
