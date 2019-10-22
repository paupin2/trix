package trix

import (
	"strconv"
	"strings"
)

// Reply represents a map with multiple values for each key
type Reply map[string][]string

// Set the specifued value(s) for the key
func (reply *Reply) Set(key string, value ...string) {
	(*reply)[key] = value
}

// Add the specifued value(s) to the key
func (reply *Reply) Add(key string, value ...string) {
	(*reply)[key] = append((*reply)[key], value...)
}

// Errors returns a list of all key/values that look like an error
func (reply *Reply) Errors() Reply {
	errors := Reply{}
	for key, values := range *reply {
		for _, value := range values {
			if value[:6] == "ERROR_" {
				errors[key] = append(errors[key], value)
			}
		}
	}
	return errors
}

// ErrorReason returns a simple string with the reason a transaction failed.
func (reply *Reply) ErrorReason() string {
	if (*reply)["status"][0] == "TRANS_OK" {
		// no error
		return ""
	}

	for _, values := range reply.Errors() {
		for _, value := range values {
			if value[:6] == "ERROR_" {
				return value
			}
		}
	}
	return "TRANS_ERROR"
}

// Get returns the first value of a the reply's key
func (reply Reply) Get(key string) string {
	if values := reply[key]; len(values) > 0 {
		return values[0]
	}
	return ""
}

// GetInt returns the first value of a the reply's key, as an int
func (reply Reply) GetInt(key string) int {
	if i, err := strconv.Atoi(reply.Get(key)); err == nil {
		return i
	}
	return 0
}

// GetBool returns the first value of a the reply's key, as a bool
func (reply Reply) GetBool(key string) bool {
	switch strings.ToLower(reply.Get(key)) {
	case "1", "t", "true", "on":
		return true
	}
	return false
}
