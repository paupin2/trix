package trix

// GetSettings returns the settings values that matches the environment,
// starting from the matched nodes. It should be called with a spec matching
// the nodes where settings should be run, and usually a temporary environment
// (including both the configuration and the current keys) will be created
// for the function call.
//
// The settings structure is a flexible way to store and parse configuration.
// In this structure, we define a set of keys within a node, and these keys
// are then used to find values using a key lookup function.
// However, with flexibility comes complexity.
//
// Let's say we want to store configuration about how to display a label for
// the zipcode field, on an classified ad, and there are multiple possibilities:
//
// 1. In any case, we want to use the label "Zip code"
//
// 2. If the category is "1001", and the ad is for a sale, we want to add
// the suffix "(of house)", because the zipcode refers to a house.
//
// 3. On the "1002", and if the ad is for renting, we want to add the
// suffix "(of apartment)".
//
// 4. On any category, if the field "pickup_location" is present, we want to
// show the suffix "(of pick-up location)".
//
// On this example, using the following configuration:
//
//   settings.1.default=label:Zip code
//   settings.1.continue=1
//   settings.2.keys.1=category
//   settings.2.keys.2=type
//   settings.2.1001.sale.value=suffix:(of house)
//   settings.2.1002.rent.value=suffix:(of apartment)
//   settings.3.keys.1=?pickup_location
//   settings.3.true.value=suffix:(of pick-up location)
//
// Settings are evaluated like a series of cases in a switch statement. The
// first matching case breaks the switch, unless "continue=1" is used.
// Each matching case can define one or more values, and the same key can be
// defined more than once, and all values are returned.
//
// If no key is used, "value" is assumed.
func (node *Node) GetSettings(keys ...interface{}) Reply {
	reply := Reply{}
	if node == nil || len(keys) < 1 {
		// avoid a segfault
		return reply
	}

	usePrefix := false
	prefix := ""
	parsealue := func(value string) {
		for _, value := range splitEsc(value, ",", `\`) {
			var subKey, subValue string
			if parts := splitNEsc(value, ":", `\`, 2); len(parts) == 2 {
				subKey, subValue = parts[0], parts[1]
			} else {
				subKey, subValue = "value", parts[0]
			}

			if usePrefix {
				if subKey == "value" {
					subKey = prefix
				} else {
					subKey = prefix + "_" + subKey
				}
			}
			reply[subKey] = append(reply[subKey], subValue)
		}
	}

	// if we're returning multiple settings, prefix each one with the parent
	// settings root node's key, followed by an underscore.
	if strKeys := ParseKeys(keys); strKeys[len(strKeys)-1] == "*" {
		usePrefix = true
	}

	// for each node matching the spec, run settings on it
	for _, settingNode := range node.GetNodes(keys...) {
		// each setting may have multiple cases, that are evaluated in order.
		// the first matching case is returned; unless the case node has a
		// `continue=1` key, matching stops after the first match.
		if usePrefix {
			prefix = settingNode.Key
		}

		for _, caseNode := range settingNode.GetNodes("*") {
			matched := false
			if defaultNode := caseNode.GetNode("default"); defaultNode != nil {
				// the `default` node takes precedence over others;
				// if it's present, use its value
				parsealue(defaultNode.internalStringValue())
				matched = true

			} else if keysNode := caseNode.GetNode("keys"); keysNode != nil {
				// next try matching the values for the `keys` node.
				wantedKeys := keysNode.GetStringValues("*")
				valueSpec := make([]interface{}, len(wantedKeys)+1)
				for i := 0; i < len(wantedKeys); i++ {
					if key := wantedKeys[i]; key[0] == '?' {
						// when the key name starts with '?', instead of the
						// key's value, use "true" if the key is present or
						// "false" otherwise.
						key = key[1:]
						if _, err := node.TryGet(key); err == nil {
							valueSpec[i] = "true"
						} else {
							valueSpec[i] = "false"
						}
					} else {
						valueSpec[i] = node.Get(key)
					}
				}
				valueSpec[len(wantedKeys)] = "value"

				if valueNode := caseNode.GetNode(valueSpec...); valueNode != nil {
					matched = true
					parsealue(valueNode.internalStringValue())
				}
			}

			if matched && !caseNode.GetBool("continue") {
				break
			}
		}
	}
	return reply
}
