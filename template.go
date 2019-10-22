package trix

import (
	"html/template"
)

// TemplateFuncs returns an map suitable as an argument to template.Funcs.
// This map contains some useful functions to use trix inside Go templates.
// The values com from this node.
func (node *Node) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"get": func(keys ...interface{}) Value {
			return node.Get(keys...)
		},
		"getnodes": func(keys ...interface{}) NodeList {
			return node.GetNodes(keys...)
		},
		"getvalues": func(keys ...interface{}) []Value {
			return node.GetValues(keys...)
		},
		"getmap": func(keys ...interface{}) Args {
			return node.GetMap(keys...)
		},
		"getsettings": func(keys ...interface{}) Reply {
			return node.GetSettings(keys...)
		},
	}
}
