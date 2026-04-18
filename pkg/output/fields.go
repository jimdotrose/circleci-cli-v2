package output

import (
	"reflect"
	"strings"
)

// JSONFields returns the JSON field names for a struct type by inspecting
// `json` struct tags. If v is a pointer or slice, the element type is used.
// Embedded structs are flattened. Fields tagged `json:"-"` are excluded.
func JSONFields(v interface{}) []string {
	t := reflect.TypeOf(v)
	for t != nil && (t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice) {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return nil
	}
	return collectFields(t)
}

func collectFields(t reflect.Type) []string {
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			fields = append(fields, collectFields(f.Type)...)
			continue
		}
		tag := f.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.SplitN(tag, ",", 2)[0]
		if name != "" && name != "-" {
			fields = append(fields, name)
		}
	}
	return fields
}
