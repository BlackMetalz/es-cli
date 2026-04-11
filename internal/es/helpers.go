package es

import "fmt"

// JsonStr safely converts an interface{} value to a string.
// Returns empty string for nil values.
func JsonStr(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
