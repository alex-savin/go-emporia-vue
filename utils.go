package vue

import (
	"reflect"
	"strconv"
	"strings"
)

// isNil checks if an interface value is nil, including typed nil values
// such as nil pointers, maps, arrays, channels, and slices.
func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Pointer, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

// intArrayToString converts a slice of integers to a comma-separated string.
func intArrayToString(a []int) string {
	var IDs []string
	for _, i := range a {
		IDs = append(IDs, strconv.Itoa(i))
	}

	return strings.Join(IDs, ",")
}
