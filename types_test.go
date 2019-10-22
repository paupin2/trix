package trix

import (
	"testing"
)

func TestNumericStringSlice(t *testing.T) {
	s := NumericStringSlice{"a", "0", "a1", "lol", "3", "99", "00", "03", "000"}
	s.Sort()
	// "0", "00", "a" or any other non-numeric string are considered equal
	testEqualString(t, s, "[a 0 a1 lol 00 000 3 03 99]")
}

func TestArgs(t *testing.T) {
	a := Args{"a": 1}
	b := Args{"b": 2}
	testEqualString(t, a, `args[a:1]`)
	testEqualString(t, b, `args[b:2]`)

	c := a.Clone()
	a.Merge(b)
	testEqualString(t, a, `args[a:1 b:2]`)
	testEqualString(t, c, `args[a:1]`) // the clone is unchanged
	testEqualString(t, b, `args[b:2]`)

	d := Args{"d": 4}
	testEqualString(t, d, `args[d:4]`)

	e := d.Add(a)
	testEqualString(t, a, `args[a:1 b:2]`)
	testEqualString(t, d, `args[d:4]`) // c is unchanged
	testEqualString(t, e, `args[a:1 b:2 d:4]`)

	f := Args{"int": 1, "bool": true, "str": "a"}
	testEqualString(t, f.GetString("int"), "1")
	testEqualString(t, f.GetString("bool"), "true")
	testEqualString(t, f.GetString("str"), "a")
}
