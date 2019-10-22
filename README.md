# trix - A tree data structure that can do many tricks

Trix is a purely-go, sorted, stackable tree data structure.
It's a multi-level map, that's especially useful for storing and
querying configuration, and for building and outputting complex JSON data.

## Install

	go get github.com/paupin2/trix

## Docs

A big effort was done do adequately document the code,
so you can either browse it directly or do `godoc -http=:6060` and then visit:
http://localhost:6060/pkg/github.com/paupin2/trix/

## Tests

You can run `go test` directly, or run `./test.sh`, which will also fmt/vet.

## Features

* Stackable - scopes can be inherited
* Easily load from files - parsing `include` directives and comments
* Serialize and parse JSON and plain text - keeping data types
* Easily find, transverse and change multiple items

# What is it good for?

I'm glad you asked! First, let's add a few entries:

```go
// Initialize a new root, add a few nodes
conf := trix.NewRoot()
conf.MergeReader(bytes.NewBufferString(`
	many.levels.deep.key=value
	server.timeout=10s
	url.base=https://coolstore.com
	sales.vat=.23
	item.1.price=10
	item.1.name=Socks
	item.2.price=25
	item.2.name=Cool shirt
	item.3.price=17
	item.3.name=Coffee mug
`), true)
```

There are many ways to populate a Trix tree. The most common way is to use
`MergeFile` to read configuration files, which can themselves include other
configuration files. You can also read entries from an `io.Reader`
(as we do above), or add entries manually.

Now we can fetch some values. Trix is very useful for storing configuration.
When reading from a buffer, the expected format is `key=value`, where `key`
can have multiple (dot-deparated) levels. The values are usually loaded as
strings, but it's possible to convert them when reading.

```go
p := func(v interface{}) { fmt.Println(reflect.TypeOf(v), v) }
p(conf.GetStringDefault("localhost:1234", "server.address")) // string localhost:1234
p(conf.GetDuration("server.timeout"))                        // time.Duration 10s
p(conf.GetValues("item.*.price"))                            // []trix.Value [10 25 17]
p(strings.Join(conf.GetStringValues("item.*.name"), ", "))   // string Socks, Cool shirt, Coffee mug

describe := func(node *trix.Node) trix.Value {
	return fmt.Sprintf("%s (%.02f €)",
		node.Get("name"),
		node.GetFloat("price")*(1+conf.GetFloat("sales.vat")),
	)
}
p(conf.GetNodes("item.*").ForEach(describe)) // []trix.Value [Socks (12.30 €) Cool shirt (30.75 €) Coffee mug (20.91 €)]
```

Though values are initially read as strings, they are stored as `interface{}`.
Functions are provided to convert values to the most commonly-used types
(int, float, string, bool, duration, time), but as long as you cast the values,
you can so you can store anything, really. Here is an overly-complicated example
where we store a closure. For no apparent reason.

```go
conf.SetKey("fx.addvat", func() func(v float64) trix.Value {
	vat := 1.23
	return func(v float64) trix.Value {
		return v * vat
	}
}())
vatter := conf.Get("fx.addvat").(func(v float64) trix.Value)
p(vatter(11)) // float64 13.53
```

Trix is also very useful in templates. The `TemplateFuncs` method will return
functions suitable for inserting into regular templates. By using `With()`
you can create child scopes, which are useful to mix configuration with
local variables, for instance.

```go
// create a new scope on top of `conf`
scope := conf.With(trix.Args{
	"name":  "Mr. T",
	"quote": "I pity the fool",
})
tpl, _ := template.New("home").
	Funcs(scope.TemplateFuncs()).
	Funcs(template.FuncMap{"describe": describe}).
	Parse(`
	<p><cite>{{ get "quote" }}</cite> by {{ get "name" }}</p>
<ul>
	{{ range (getnodes "item.*") }}
		<li>
			<a href="{{ get "url.base" }}?id={{ .Key }}">{{ describe . }}</a>
		</li>
	{{ end }}
</ul>
`)
tpl.Execute(os.Stdout, "")
// <p><cite>I pity the fool</cite> by Mr. T</p>
// <ul>
//   <li>
//     <a href="https://coolstore.com?id=1">Socks (12.30 €)</a>
//   </li>
//   <li>
//     <a href="https://coolstore.com?id=2">Cool shirt (30.75 €)</a>
//   </li>
//   <li>
//     <a href="https://coolstore.com?id=3">Coffee mug (20.91 €)</a>
//   </li>
// </ul>
```

There are also quite a few other uses, which I hope you'll get to find out
for yourself.

In addition to all of that, Trix also is an essential part of a balanced diet.

# Development and contributing

Pull requests and issues are always welcome.
Please use `./test.sh` before send pull requests, to ensure fmt/vet/test are OK.
