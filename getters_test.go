package trix

import (
	"testing"
	"time"
)

func TestTryGetNode(t *testing.T) {
	var node *Node
	shouldFail := func(expected string, keys ...interface{}) {
		t.Helper()
		_, err := node.TryGetNode(keys...)
		testError(t, err, expected)
	}

	// node == nil, but this should not segfault
	shouldFail("node not found", "x.y")

	node = NewNode("lol")
	shouldFail("node not found", "x.y")

	node.SetKey("x.y", "a")
	_, err := node.TryGetInt("x.y")
	testError(t, err, `strconv.ParseInt: parsing "a": invalid syntax`)

	_, err = node.TryGetFloat("x.y")
	testError(t, err, `strconv.ParseFloat: parsing "a": invalid syntax`)

	_, err = node.TryGetDuration("x.y")
	testError(t, err, `bad duration`)

	_, err = node.TryGetBool("x.y")
	testError(t, err, `bad value`)

	node.SetKey("x.a", "true")
	testDeepEqual(t, node.GetBool("x.a"), true)

	node.SetKey("x.b", "17")
	testDeepEqual(t, node.GetInt("x.b"), 17)

	node.SetKey("x.c", "2d1h20m")
	testDeepEqual(t, node.GetDuration("x.c"), time.Hour*49+time.Minute*20)
}

func TestIterate(t *testing.T) {
	de := NewRoot()
	de.SetKey("de.2", "zwei")
	de.SetKey("de.1", "eins")
	de.SortRecursively()

	fr := de.With()
	fr.SetKey("fr.1", "un")
	fr.SetKey("fr.2", "deux")
	fr.SetKey("fr.3", "trois")
	fr.SortRecursively()

	en := fr.With()
	en.SetKey("en.1", "one")
	en.SetKey("en.2", "two")
	en.SetKey("en.3", "three")
	en.SetKey("en.4", "four")
	en.SetKey("en.4", "five")
	en.SortRecursively()

	// sort order
	testDeepEqual(t, de.GetValues("de.*"), []Value{"eins", "zwei"})

	// order is top-scope to bottom-scope
	testDeepEqual(t, en.GetValues("*.1"), []Value{"one", "un", "eins"})
}

func TestGettersDefaults(t *testing.T) {
	root := NewRoot()
	testTrue(t, root.Get("missing") == nil)
	testTrue(t, root.GetNode("missing") == nil)
	testTrue(t, root.GetString("missing") == "")
	testTrue(t, root.GetInt("missing") == 0)
	testTrue(t, root.GetFloat("missing") == 0.0)
	testTrue(t, root.GetBool("missing") == false)
	testTrue(t, root.GetDuration("missing") == time.Duration(0))
}

func TestDefaultGetters(t *testing.T) {
	root := NewRoot()
	root.SetKey("main.key", "1")
	root.SetKey("main.duration", "10m")

	testTrue(t, root.GetNodeDefault(nil, "missing.path") == nil)
	testDeepEqual(t, root.GetNodeDefault(nil, "main"), root.GetNode("main"))
	testDeepEqual(t, root.GetDefault("hi", "missing.path"), "hi")
	testDeepEqual(t, root.GetDefault("hi", "main.key"), "1")
	testDeepEqual(t, root.GetStringDefault("x", "missing.path"), "x")
	testDeepEqual(t, root.GetStringDefault("17", "main.key"), "1")
	testDeepEqual(t, root.GetIntDefault(17, "missing.path"), 17)
	testDeepEqual(t, root.GetIntDefault(17, "main.key"), 1)
	testDeepEqual(t, root.GetFloatDefault(17.0, "missing.path"), 17.0)
	testDeepEqual(t, root.GetFloatDefault(17, "main.key"), 1.0)
	testDeepEqual(t, root.GetBoolDefault(true, "missing.path"), true)
	testDeepEqual(t, root.GetBoolDefault(false, "main.key"), true)
	testDeepEqual(t, root.GetDurationDefault(time.Minute, "missing.path"), time.Minute)
	testDeepEqual(t, root.GetDurationDefault(0, "main.duration"), time.Minute*10)
}

func TestSimpleGetters(t *testing.T) {
	root := NewRoot()
	root.SetKey("string.one", "1")
	root.SetKey("bool", "true")
	root.SetKey("float", "3.14159")
	root.SetKey("duration", "1h")

	testDeepEqual(t, root.Get("string.one"), "1")
	testEqualString(t, root.GetNode("string"), "{one=1}")
	testDeepEqual(t, root.GetString("bool"), "true")
	testDeepEqual(t, root.GetInt("string.one"), 1)
	testDeepEqual(t, root.GetFloat("float"), 3.14159)
	testDeepEqual(t, root.GetBool("bool"), true)
	testDeepEqual(t, root.GetDuration("duration"), time.Hour)
}

func TestMustGetters(t *testing.T) {
	p := func(f func()) (didItPanic bool) {
		defer func() {
			if r := recover(); r != nil {
				didItPanic = true
			}
		}()
		f()
		return
	}

	root := NewRoot()
	root.SetKey("string.one", "1")
	root.SetKey("string.two", "2")
	root.SetKey("string.three", "3")
	root.SetKey("bool.one", "true")
	root.SetKey("float.one", "3.14159")
	root.SetKey("duration.one", "1h")

	testTrue(t, p(func() { root.MustGet("missing.node") }))
	testTrue(t, !p(func() { root.MustGet("string.one") }))
	testTrue(t, p(func() { root.MustGetNode("missing.node") }))
	testTrue(t, !p(func() { root.MustGetNode("string.one") }))
	testTrue(t, p(func() { root.MustGetString("missing.node") }))
	testTrue(t, !p(func() { root.MustGetString("string.one") }))
	testTrue(t, p(func() { root.MustGetInt("missing.node") }))
	testTrue(t, p(func() { root.MustGetInt("bool.one") })) // invalid int also panics
	testTrue(t, !p(func() { root.MustGetInt("string.one") }))
	testTrue(t, p(func() { root.MustGetFloat("missing.node") }))
	testTrue(t, p(func() { root.MustGetFloat("bool.one") })) // invalid float also panics
	testTrue(t, !p(func() { root.MustGetFloat("float.one") }))
	testTrue(t, !p(func() { root.MustGetFloat("string.one") }))
	testTrue(t, p(func() { root.MustGetBool("missing.node") }))
	testTrue(t, p(func() { root.MustGetBool("string.two") })) // invalid bool also panics
	testTrue(t, !p(func() { root.MustGetBool("string.one") }))
	testTrue(t, p(func() { root.MustGetDuration("missing.node") }))
	testTrue(t, p(func() { root.MustGetDuration("string.one") })) // invalid duration also panics
	testTrue(t, !p(func() { root.MustGetDuration("duration.one") }))
}

func TestExtraGetters(t *testing.T) {
	root := NewRoot()
	root.SetKey("main.string.one", "1")
	root.SetKey("main.string.two", "2")
	root.SetKey("main.string.three", "3")
	root.SetKey("main.bool.one", "true")
	root.SetKey("main.bool.two", "false")
	root.SetKey("main.duration.one", "1h")
	root.SetKey("main.duration.two", "1m10s")
	root.SetKey("*.star", "2")

	testDeepEqual(t, root.GetValues("main.*.one"), []Value{"1", "true", "1h"})
	testDeepEqual(t, root.GetValues("main.string.*"), []Value{"1", "2", "3"})
	testDeepEqual(t, root.GetValues("*.star"), []Value{"2"})
	testDeepEqual(t, root.GetValues("*.*"), []Value{"2"})
	testDeepEqual(t, root.GetMap("main.*.one"), Args{
		"string":   "1",
		"bool":     "true",
		"duration": "1h",
	})
	testDeepEqual(t, root.GetStringMap("main.*.one"), StrArgs{
		"string":   "1",
		"bool":     "true",
		"duration": "1h",
	})
	testDeepEqual(t, root.GetValues("main.*.one"), []Value{"1", "true", "1h"})

}

func TestPreventSegfault(t *testing.T) {
	testTrue(t, (*Node)(nil).GetNode("missing.key") == nil)
}
