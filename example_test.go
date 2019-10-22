package trix_test

import (
	"fmt"
	"html/template"
	"os"
	"reflect"
	"sort"

	"github.com/paupin2/trix"
)

func Example() {
	// Initialize a new root, add a few nodes
	first := trix.NewRoot()
	first.SetKey("main.string.one", "1")
	first.SetKey("main.1.one", 1)
	first.SetKey("main.bool.one", true)

	fmt.Println(first.GetValues("main.*.one")) // wildcard matching

	second := first.With(trix.Args{
		"main.string.one": "one",
		"main.string.two": "two",
	})
	fmt.Println(second.Get("main.bool.one"))   // inherit from first
	fmt.Println(second.Get("main.string.one")) // get from second
	fmt.Println(second.Get("main", 1, "one"))  // each argument is converted to a string
	// Output:
	// [1 1 true]
	// true
	// one
	// 1
}

func ExampleNode_Get() {
	// Initialize a new root, add a few nodes
	root := trix.FromArgs(trix.Args{"m.int": 1})
	p := func(v interface{}) {
		fmt.Println(reflect.TypeOf(v), v)
	}

	p(root.Get("m.int")) // returns the actual value, whatever its type
	p(root.GetNode("m"))
	p(root.GetInt("m.int"))      // converts to int if necessary
	p(root.GetFloat("m.int"))    // converts to float if necessary
	p(root.GetString("m.int"))   // converts to string if necessary
	p(root.GetBool("m.int"))     // converts to bool (strings 1/t/true/on == true)
	p(root.GetDuration("m.int")) // error parsing, use default type value

	// Output:
	// int 1
	// *trix.Node {int=1}
	// int 1
	// float64 1
	// string 1
	// bool true
	// time.Duration 0s
}

func ExampleNode_TryGet() {
	// Initialize a new root, add a few nodes
	root := trix.FromArgs(trix.Args{"m.int": 1})
	p := func(v interface{}, err error) {
		fmt.Println(reflect.TypeOf(v), v, "|", err)
	}

	p(root.TryGet("m.int")) // returns the actual value, whatever its type
	p(root.TryGetNode("m"))
	p(root.TryGetNode("missing.node"))
	p(root.TryGetInt("m.int"))    // converts to int if necessary
	p(root.TryGetFloat("m.int"))  // converts to float if necessary
	p(root.TryGetString("m.int")) // converts to string if necessary
	p(root.TryGetBool("m.int"))   // converts to bool (strings 1/t/true/on == true)
	p(root.TryGetDuration("m.int"))

	// Output:
	// int 1 | <nil>
	// *trix.Node {int=1} | <nil>
	// *trix.Node  | node not found
	// int 1 | <nil>
	// float64 1 | <nil>
	// string 1 | <nil>
	// bool true | <nil>
	// time.Duration 0s | bad duration
}

func ExampleNode_GetMap() {
	// Initialize a new root, add a few nodes
	root := trix.NewRoot()
	root.AddNode("catg").Push().MergeArgs(trix.Args{"id": 10, "name": "News"})
	root.AddNode("catg").Push().MergeArgs(trix.Args{"id": 10, "name": "Stories"})
	root.AddNode("catg").Push().MergeArgs(trix.Args{"id": 10, "name": "Opinion pieces"})
	p := func(v interface{}) {
		fmt.Println(reflect.TypeOf(v), v)
	}
	parse := func(node *trix.Node) trix.Value {
		return fmt.Sprintf("%v:%v", node.Get("id"), node.Get("name"))
	}

	p(root.GetMap("catg.*.name"))
	p(root.GetNodes("catg.*").ForEach(parse))
	p(root.GetValues("catg.*.name"))

	// Output:
	// trix.Args args[1:News 2:Stories 3:Opinion pieces]
	// []trix.Value [10:News 10:Stories 10:Opinion pieces]
	// []trix.Value [News Stories Opinion pieces]
}

func ExampleNode_GetStringMap() {
	// Initialize a new root, add a few nodes
	root := trix.NewRoot()
	root.SetKey("favourite.numbers.1", 2*5)
	root.SetKey("favourite.numbers.2", "ten")
	root.SetKey("favourite.numbers.3", 3.14)
	p := func(v interface{}) {
		fmt.Println(reflect.TypeOf(v), v)
	}

	p(root.GetStringValues("favourite.numbers.*"))
	// Output:
	// []string [10 ten 3.14]
}

func ExampleNode_GetSettings() {
	conf := trix.NewRoot()
	sett := conf.AddNode("settings")
	sett.SetKey("1.default", "label:Zip code")
	sett.SetKey("1.continue", "1")
	sett.SetKey("2.keys.1", "category")
	sett.SetKey("2.keys.2", "type")
	sett.SetKey("2.3041.s.value", "suffix:(of house)")
	sett.SetKey("2.3042.u.value", "suffix:(of apartment)")
	sett.SetKey("3.keys.1", "?pickup_location")
	sett.SetKey("3.true.value", "suffix:(of pick-up location)")

	p := func(r trix.Reply) {
		// print nicely-sorted reply
		keys := make([]string, 0, len(r))
		for k := range r {
			keys = append(keys, k)
		}
		sort.StringSlice(keys).Sort()
		for i, k := range keys {
			if i > 0 {
				fmt.Printf(" ")
			}
			fmt.Printf("%s:%v", k, r[k])
		}
		fmt.Printf("\n")
	}

	p(conf.With(trix.Args{"category": 3041}).GetSettings("settings"))
	p(conf.With(trix.Args{"category": 3041, "type": "s"}).GetSettings("settings"))
	p(conf.With(trix.Args{"pickup_location": "whatever"}).GetSettings("settings"))
	p(conf.GetSettings("settings"))

	// Output:
	// label:[Zip code]
	// label:[Zip code] suffix:[(of house)]
	// label:[Zip code] suffix:[(of pick-up location)]
	// label:[Zip code]
}

func ExampleNode_TemplateFuncs() {
	// Initialize a new root, add a few nodes
	t := trix.NewRoot()
	t.SetKey("url.base", "http://example.com")
	t.AddNode("item").Push().MergeArgs(trix.Args{"id": "jhn", "name": "John"})
	t.GetNode("item").Push().MergeArgs(trix.Args{"id": "mry", "name": "Mary"})

	tpl, _ := template.New("home").
		Funcs(t.TemplateFuncs()).
		Parse(`
		<ul>
			{{ range (t_getnodes "item.*") }}
				<li>
					<a href="{{ t_get "url.base" }}/{{.Get "id"}}">{{ .Get "name" }}</a>
				</li>
			{{ end }}
		</ul>
	`)
	tpl.Execute(os.Stdout, "")

	// Output:
	// <ul>
	// <li>
	// <a href="http://example.com/jhn">John</a>
	// </li>
	// <li>
	// <a href="http://example.com/mry">Mary</a>
	// </li>

	// </ul>
}
