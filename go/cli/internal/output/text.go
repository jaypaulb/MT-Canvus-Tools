package output

import (
	"fmt"
	"reflect"
	"strings"
)

// TextFormatter formats data as simple text output.
// For single objects: outputs as key-value pairs.
// For lists: outputs as simple newline-separated values.
type TextFormatter struct {
	// Fields specifies which fields to display for list output
	// If empty, a default field (like "id" or "name") is used
	Fields []string
}

// Format converts the input data to simple text format.
func (f *TextFormatter) Format(data interface{}) (string, error) {
	if data == nil {
		return "", nil
	}

	// Handle different data types
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return f.formatList(data), nil
	case reflect.Map:
		return f.formatKeyValue(data), nil
	case reflect.Struct:
		return f.formatKeyValue(data), nil
	default:
		// For simple types, just convert to string
		return fmt.Sprintf("%v", data), nil
	}
}

// formatList formats a slice as newline-separated values
func (f *TextFormatter) formatList(data interface{}) string {
	v := reflect.ValueOf(data)
	if v.Len() == 0 {
		return ""
	}

	var lines []string
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface()
		line := f.formatListItem(item)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// formatListItem formats a single item for list output
func (f *TextFormatter) formatListItem(data interface{}) string {
	// If specific fields are requested, output them
	if len(f.Fields) > 0 {
		values := make([]string, 0, len(f.Fields))
		for _, field := range f.Fields {
			value := f.getValue(data, field)
			values = append(values, fmt.Sprintf("%v", value))
		}
		return strings.Join(values, " ")
	}

	// Otherwise, try to find a meaningful field (id, name, or first field)
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Map {
		// Try common field names
		for _, key := range []string{"id", "name", "title"} {
			for _, mapKey := range v.MapKeys() {
				if strings.ToLower(fmt.Sprintf("%v", mapKey.Interface())) == key {
					return fmt.Sprintf("%v", v.MapIndex(mapKey).Interface())
				}
			}
		}
		// Fall back to first key
		if v.Len() > 0 {
			firstKey := v.MapKeys()[0]
			return fmt.Sprintf("%v", v.MapIndex(firstKey).Interface())
		}
	}

	// For structs or other types, use string representation
	return fmt.Sprintf("%v", data)
}

// formatKeyValue formats data as key-value pairs
func (f *TextFormatter) formatKeyValue(data interface{}) string {
	var lines []string

	v := reflect.ValueOf(data)

	// Handle map
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			value := v.MapIndex(key).Interface()
			lines = append(lines, fmt.Sprintf("%s: %v", keyStr, value))
		}
		return strings.Join(lines, "\n")
	}

	// Handle struct
	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i).Interface()

			// Use yaml tag if available, otherwise field name
			fieldName := field.Name
			tag := field.Tag.Get("yaml")
			if tag != "" && tag != "-" {
				fieldName = strings.Split(tag, ",")[0]
			}

			lines = append(lines, fmt.Sprintf("%s: %v", fieldName, value))
		}
		return strings.Join(lines, "\n")
	}

	return fmt.Sprintf("%v", data)
}

// getValue extracts a value from data by field name
func (f *TextFormatter) getValue(data interface{}, field string) interface{} {
	v := reflect.ValueOf(data)

	// Handle map
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			if strings.ToLower(fmt.Sprintf("%v", key.Interface())) == strings.ToLower(field) {
				return v.MapIndex(key).Interface()
			}
		}
		return ""
	}

	// Handle struct
	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			fieldType := t.Field(i)
			// Check yaml tag or field name
			tag := fieldType.Tag.Get("yaml")
			fieldName := fieldType.Name
			if tag != "" {
				fieldName = strings.Split(tag, ",")[0]
			}
			if strings.ToLower(fieldName) == strings.ToLower(field) {
				return v.Field(i).Interface()
			}
		}
	}

	return ""
}
