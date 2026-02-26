// Package output provides JSON formatting utilities for CLI output.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

// PrintJSON marshals v to compact JSON (pruning nil/empty/zero fields) with 2-space indent and prints to stdout.
func PrintJSON(v any) {
	pruned := prune(v)
	data, err := json.MarshalIndent(pruned, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"error":"%s"}`+"\n", err.Error())
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// PrintError prints a JSON error to stderr and exits with code 1.
func PrintError(err error) {
	fmt.Fprintf(os.Stderr, `{"error":"%s"}`+"\n", err.Error())
	os.Exit(1)
}

// prune recursively removes nil, empty, and zero-value fields from maps and slices.
func prune(v any) any {
	if v == nil {
		return nil
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map:
		out := make(map[string]any)
		iter := rv.MapRange()
		for iter.Next() {
			key := fmt.Sprintf("%v", iter.Key().Interface())
			val := prune(iter.Value().Interface())
			if val != nil {
				out[key] = val
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out

	case reflect.Slice, reflect.Array:
		if rv.Len() == 0 {
			return nil
		}
		out := make([]any, 0, rv.Len())
		for i := range rv.Len() {
			val := prune(rv.Index(i).Interface())
			if val != nil {
				out = append(out, val)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out

	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return nil
		}
		return prune(rv.Elem().Interface())

	case reflect.Struct:
		out := make(map[string]any)
		rt := rv.Type()
		for i := range rv.NumField() {
			field := rt.Field(i)
			if !field.IsExported() {
				continue
			}
			tag := field.Tag.Get("json")
			if tag == "-" {
				continue
			}
			name := field.Name
			omitempty := false
			if tag != "" {
				parts := splitTag(tag)
				if parts[0] != "" {
					name = parts[0]
				}
				for _, p := range parts[1:] {
					if p == "omitempty" {
						omitempty = true
					}
				}
			}
			val := rv.Field(i)
			if omitempty && val.IsZero() {
				continue
			}
			pruned := prune(val.Interface())
			if pruned != nil {
				out[name] = pruned
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out

	case reflect.String:
		if rv.String() == "" {
			return nil
		}
		return v

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rv.Int() == 0 {
			return nil
		}
		return v

	case reflect.Float32, reflect.Float64:
		if rv.Float() == 0 {
			return nil
		}
		return v

	case reflect.Bool:
		if !rv.Bool() {
			return nil
		}
		return v

	default:
		return v
	}
}

func splitTag(tag string) []string {
	var parts []string
	current := ""
	for _, c := range tag {
		if c == ',' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	parts = append(parts, current)
	return parts
}
