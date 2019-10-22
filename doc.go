// Package trix is a stackable tree data structure that has many different uses.
//
// A tree structure is similar to a (sorted) map, where each value can also
// contain another map, an so on.
//
// It's also stackable, that is, a second tree may be stacked on top of a
// a previous one, like a new scope/closure. When fetching values, if the key
// is not found on the top-most tree, the previous one is searched, an so on.
//
// To find nodes, a simple path is used, with keys for successive levels
// separated by dots. All functions that accept a path use varargs and accept
// keys from any type, so that it's not necessary to convert each key to a
// string beforehand.
//
// Wildcards are also accepted when looking for multiple nodes, through the
// special key asterisk ("*").
//
// There are multiple ways to access node values, but they're mostly divided
// in four groups:
//
// 1. "Normal" getters: Get, GetNode, GetString, GetInt, GetFloat, GetBool
// and GetDuration.
//
// These will return either the plain value of the node (Get), or the value
// converted to some type.
// If the value is not found, or cannot be converted  to the specified type,
// the type's default value is returned instead.
//
// 2. "Try" getters: TryGet, TryGetNode, TryGetString, TryGetInt, TryGetFloat,
// TryGetBool and TryGetDuration.
//
// These will return an error value, in adition to the first one.
// If the value is not found, or cannot be converted to the specified type,
// the error description will be provided. Otherwise, the error will be nil.
//
// 3. "Default" getters: GetDefault, GetNodeDefault, GetStringDefault,
// GetIntDefault, GetFloatDefault, GetBoolDefault and GetDurationDefault.
//
// These will accept a default value as the first parameter,
// and return it in case something goes wrong.
//
// 4. "Extra" getters: GetMap, GetStringMap, GetStringValues, GetNodes,
// GetSettings and GetValues.
//
package trix
