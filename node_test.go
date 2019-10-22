package trix

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestChangeKey(t *testing.T) {
	root := NewRoot()
	root.SetKey("main.1", "one")
	root.SetKey("main.2", "two")
	root.SetKey("main.3", "three")

	testEqualString(t, root, `{main={1=one,2=two,3=three}}`)
	root.GetNode("main.2").Rename("two")
	testEqualString(t, root, `{main={1=one,3=three,two=two}}`)
}

func TestParseKeys(t *testing.T) {
	testDeepEqual(t, ParseKeys([]interface{}{
		"a",
		1,
		true,
		3.5,
	}), []string{
		"a",
		"1",
		"true",
		"3",
		"5",
	})
}

func TestDepth(t *testing.T) {
	root := NewRoot()
	testDeepEqual(t, root.Depth(), 0)
	testDeepEqual(t, root.AddNode("subnode").Depth(), 1)
	testDeepEqual(t, root.AddNode("really.deep.subnode").Depth(), 3)
	testDeepEqual(t, root.AddNode("really.really.deep.subnode").Depth(), 4)

	original := NewRoot()
	new := original.With()
	testDeepEqual(t, new.Depth(), 0) // even though new.Parent != nil
	testDeepEqual(t, new.AddNode("subnode").Depth(), 1)
	testDeepEqual(t, new.AddNode("really.deep.subnode").Depth(), 3)
	testDeepEqual(t, new.AddNode("really.really.deep.subnode").Depth(), 4)
}

func TestMerge(t *testing.T) {
	root1 := NewRoot()
	root2 := NewRoot()
	root3 := NewRoot()
	node := root1.SetKey("main.child.point", "value")

	testEqualString(t, root1, "{main={child={point=value}}}")
	testEqualString(t, root2, "{}")
	testEqualString(t, root3, "{}")
	testDeepEqual(t, node.Parent, root1.GetNode("main.child"))
	testDeepEqual(t, node.Depth(), 3)

	// merge (clone)
	root2.Merge(node)
	testEqualString(t, root1, "{main={child={point=value}}}")
	testEqualString(t, root2, "{point=value}")
	testEqualString(t, root3, "{}")

	// adopt (move)
	root3.Adopt(node)
	testDeepEqual(t, node.Depth(), 1)
	testDeepEqual(t, node.Parent, root3)
	testEqualString(t, root1, "{main={child=}}")
	testEqualString(t, root2, "{point=value}")
	testEqualString(t, root3, "{point=value}")
}

func TestPush(t *testing.T) {
	root1 := NewRoot()
	root1.SetKey("settings.1.default", "label:Zip code")
	root1.SetKey("settings.1.continue", "1")
	root1.SetKey("settings.2.keys.1", "category")
	root1.SetKey("settings.2.keys.2", "type")
	root1.SetKey("settings.2.3041.s.value", "suffix:(of house)")
	root1.SetKey("settings.2.3042.u.value", "suffix:(of apartment)")
	root1.SetKey("settings.3.keys.1", "?pickup_location")
	root1.SetKey("settings.3.true.value", "suffix:(of pick-up location)")
	root1.SortRecursively()

	root2 := NewRoot()
	sett := root2.AddNode("settings")
	sett.Push().MergeArgs(Args{"default": "label:Zip code", "continue": 1})
	case2 := sett.Push()
	case2.AddNode("keys").PushValues("category", "type")
	case2.SetKey("3041.s.value", "suffix:(of house)")
	case2.SetKey("3042.u.value", "suffix:(of apartment)")
	case3 := sett.Push()
	case3.SetKey("keys.1", "?pickup_location")
	case3.SetKey("true.value", "suffix:(of pick-up location)")
	root2.SortRecursively()

	testEqualString(t, root1, root2)
}

func TestPath(t *testing.T) {
	root := NewRoot()
	k := root.SetKey("settings.2.3041.s.value", "suffix:(of house)")
	testDeepEqual(t, k.Path(), []string{"settings", "2", "3041", "s", "value"})
}

func TestSort(t *testing.T) {
	items := [][]string{
		{"many.levels.deep.key", "value"},
		{"server.timeout", "10s"},
		{"sales.vat", ".23"},
		{"item.30.name", "Thirty"},
		{"item.001", "a"},
		{"item.01", "b"},
		{"item.1.name", "Socks"},
		{"item.002", "25"},
		{"item.0020.name", "Cool shirt"},
		{"item.3.name", "Coffee mug"},
	}
	// repeat a few times
	for i := 0; i < 100; i++ {
		// create new root, add items in random order
		root := NewRoot()
		for i := range rand.Perm(len(items)) {
			item := items[i]
			root.SetKey(item[0], item[1])
		}

		// sort root, dump content, compare sorted nodes
		root.SortRecursively()
		buf := bytes.Buffer{}
		root.Dump(&buf, false)
		testEqualString(t,
			strings.Split(buf.String(), "\n"),
			[]string{
				// 1
				"item.001=a",
				"item.01=b",
				"item.1.name=Socks",
				// 2
				"item.002=25",
				// 3
				"item.3.name=Coffee mug",
				// 20
				"item.0020.name=Cool shirt",
				// 30
				"item.30.name=Thirty",
				"many.levels.deep.key=value",
				"sales.vat=.23",
				"server.timeout=10s",
				"",
			},
		)
	}
}

func TestInherit(t *testing.T) {
	rootA := NewRoot()
	rootA.SetKey("main.string.one", 1)

	rootB := rootA.With()
	rootB.SetKey("main.string.three", 3)
	rootB.SetKey("main.string.four", 4)

	rootC := rootB.With()
	rootC.SetKey("main.string.three", "three")
	rootC.SetKey("main.string.five", 5)

	testDeepEqual(t, rootC.Get("main.string.one"), 1)         // inherited from A
	testDeepEqual(t, rootC.Get("main.string.four"), 4)        // inherited from B
	testDeepEqual(t, rootC.Get("main.string.three"), "three") // overwritten

	// adding values to parent trees should make them available to children
	testTrue(t, rootC.Get("main.string.two") == nil)
	rootA.SetKey("main.string.two", 2)
	testDeepEqual(t, rootC.Get("main.string.two"), 2)

	// we should get results from all contexts
	testDeepEqual(t, rootC.GetStringValues("main.*.*"), []string{"three", "5", "3", "4", "1", "2"})
}

func TestInheritGetters(t *testing.T) {
	par := NewRoot()
	par.SetKey("number.3", "three")
	par.SetKey("number.1", "one")
	par.SetKey("number.2", "two")
	par.SetKey("string.one", "1")
	par.SetKey("int.one", 1)
	par.SetKey("bool.true", true)
	par.SetKey("bool.false", false)
	par.SetKey("duration.1", time.Hour*2)
	par.SetKey("duration.2", "2m")
	par.SetKey("x.a.a", "1")
	par.SetKey("x.a.b", "2")
	par.SetKey("x.a.c", "3")
	par.SortRecursively()

	root := par.With()

	testDeepEqual(t, par.With(Args{}).GetValues("number.*"), []Value{"one", "two", "three"})

	testEqualString(t, root.GetNode("number"), `{1=one,2=two,3=three}`)
	testDeepEqual(t, root.Get("string.one"), "1")
	testDeepEqual(t, root.GetString("int.one"), "1")
	testDeepEqual(t, root.GetInt("int.one"), 1)
	testDeepEqual(t, root.GetBool("bool.true"), true)
	testDeepEqual(t, root.GetDuration("duration.1"), time.Hour*2)
	testDeepEqual(t, root.GetDuration("duration.2"), time.Minute*2)
	testDeepEqual(t, root.GetValues("number.*"), []Value{"one", "two", "three"})

	testDeepEqual(t, root.GetMap("number.*"), Args{"1": "one", "2": "two", "3": "three"})
	testDeepEqual(t, root.GetValues("number.*"), []Value{"one", "two", "three"})
	testDeepEqual(t, root.GetStringValues("number.*"), []string{"one", "two", "three"})
	testEqualString(t, root.GetNodes("x.*"), "[{a=1,b=2,c=3}]")

}

func TestAddRemove(t *testing.T) {
	S := fmt.Sprint

	root := NewRoot()
	root.SetKey("a.b.c.3", "three")
	root.SetKey("a.b.c.true", "vrai")
	testDeepEqual(t, S(root), `{a={b={c={3=three,true=vrai}}}}`)

	root.SetKey("a.b.key", "value")
	testDeepEqual(t, S(root), `{a={b={c={3=three,true=vrai},key=value}}}`)

	root.SetKey("a.b.key", "newvalue") // replace existing
	testDeepEqual(t, S(root), `{a={b={c={3=three,true=vrai},key=newvalue}}}`)

	root.GetNode("a.b").Unset("missing")
	testDeepEqual(t, S(root), `{a={b={c={3=three,true=vrai},key=newvalue}}}`) // no effect
	toRemove := root.GetNode("a.b.key")
	testTrue(t, toRemove.Parent != nil)
	testDeepEqual(t, toRemove.Depth(), 3)

	removed := root.GetNode("a.b").Unset("key")
	testDeepEqual(t, S(root), `{a={b={c={3=three,true=vrai}}}}`) // removed node
	testTrue(t, removed.Parent == nil)
	testDeepEqual(t, removed.Depth(), 0)
}

func TestSetUnset(t *testing.T) {
	par := NewRoot()
	par.SetKey("a.b.c", "old")
	root := par.With()
	root.SetKey("a.b.c", "new")

	testDeepEqual(t, root.Get("a.b.c"), "new")

	root.Unset("a.b.c")
	testDeepEqual(t, root.Get("a.b.c"), "old")
}

func TestRoot_ForEach(t *testing.T) {
	root := NewRoot()
	root.SetKey("item.1.price", "10")
	root.SetKey("item.1.name", "Socks")
	root.SetKey("item.2.price", "25")
	root.SetKey("item.2.name", "Cool shirt")
	root.SetKey("item.3.price", "17")
	root.SetKey("item.3.name", "Coffee mug")

	describe := func(node *Node) Value {
		return fmt.Sprintf("%s (%.02f €)",
			node.Get("name"),
			node.GetFloat("price")*(1+root.GetFloat("sales.vat")),
		)
	}

	// ted "[Socks (10.00 €) Cool shirt (25.00 €) Coffee mug (17.00 €)]",
	// got "[Socks (10.00 €) Cool shirt (25.00 €) Coffee mug (17.00 €)]"
	items := root.GetNodes("item.*").ForEach(describe)
	testDeepEqual(t, items, []Value{
		"Socks (10.00 €)",
		"Cool shirt (25.00 €)",
		"Coffee mug (17.00 €)",
	})
}

func TestFillKey(t *testing.T) {
	root := NewRoot()
	testEqualString(t, root, `{}`)

	root.FillKey("a", 10)
	testEqualString(t, root, `{a=10}`)

	root.FillKey("a", 20)
	testEqualString(t, root, `{a={1=10,2=20}}`)

	root.FillKey("c", 3.14)
	root.FillKey("c", "pi")
	testDeepEqual(t, root.Get("c.1"), 3.14)
}
