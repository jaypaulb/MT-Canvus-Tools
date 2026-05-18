package output

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/tabwriter"
)

// TableFormatter formats data as aligned tables with headers.
type TableFormatter struct {
	// Columns specifies the fields to display and their order
	// If empty, all fields are displayed in default order
	Columns []string
}

// Format converts the input data to a table format with aligned columns.
// Supports maps and slices of maps. Uses text/tabwriter for alignment.
func (f *TableFormatter) Format(data interface{}) (string, error) {
	if data == nil {
		return "", nil
	}

	// Handle different data types
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return f.formatSlice(data)
	case reflect.Map:
		return f.formatSingle(data)
	case reflect.Struct:
		return f.formatSingle(data)
	default:
		// For simple types, just convert to string
		return fmt.Sprintf("%v", data), nil
	}
}

// formatSlice formats a slice of items as a table
func (f *TableFormatter) formatSlice(data interface{}) (string, error) {
	v := reflect.ValueOf(data)
	if v.Len() == 0 {
		return "", nil
	}

	// Get first item to determine columns
	firstItem := v.Index(0).Interface()
	columns := f.getColumns(firstItem)
	if len(columns) == 0 {
		return "", fmt.Errorf("no columns found in data")
	}

	// Create tabwriter with 2-space padding
	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)

	// Write header
	header := strings.Join(columns, "\t")
	fmt.Fprintln(w, header)

	// Write separator
	separators := make([]string, len(columns))
	for i := range separators {
		separators[i] = strings.Repeat("-", len(columns[i]))
	}
	fmt.Fprintln(w, strings.Join(separators, "\t"))

	// Write rows
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface()
		row := f.formatRow(item, columns)
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	w.Flush()
	return buf.String(), nil
}

// formatSingle formats a single item as key-value pairs
func (f *TableFormatter) formatSingle(data interface{}) (string, error) {
	columns := f.getColumns(data)
	if len(columns) == 0 {
		return fmt.Sprintf("%v", data), nil
	}

	// Create tabwriter with 2-space padding
	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)

	// Write key-value pairs
	for _, col := range columns {
		value := f.getValue(data, col)
		fmt.Fprintf(w, "%s:\t%v\n", col, value)
	}

	w.Flush()
	return buf.String(), nil
}

// getColumns returns the list of columns to display
func (f *TableFormatter) getColumns(data interface{}) []string {
	// If columns are specified, use them
	if len(f.Columns) > 0 {
		return f.Columns
	}

	// Otherwise, extract from data
	v := reflect.ValueOf(data)

	// Handle map
	if v.Kind() == reflect.Map {
		columns := make([]string, 0, v.Len())
		for _, key := range v.MapKeys() {
			columns = append(columns, fmt.Sprintf("%v", key.Interface()))
		}
		return columns
	}

	// Handle struct
	if v.Kind() == reflect.Struct {
		t := v.Type()
		columns := make([]string, 0, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			// Use yaml tag if available, otherwise field name
			tag := field.Tag.Get("yaml")
			if tag != "" && tag != "-" {
				columns = append(columns, strings.Split(tag, ",")[0])
			} else {
				columns = append(columns, field.Name)
			}
		}
		return columns
	}

	return []string{}
}

// getValue extracts a value from data by column name
func (f *TableFormatter) getValue(data interface{}, column string) interface{} {
	v := reflect.ValueOf(data)

	// Handle map
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			if fmt.Sprintf("%v", key.Interface()) == column {
				return v.MapIndex(key).Interface()
			}
		}
		return ""
	}

	// Handle struct
	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			// Check yaml tag or field name
			tag := field.Tag.Get("yaml")
			fieldName := field.Name
			if tag != "" {
				fieldName = strings.Split(tag, ",")[0]
			}
			if fieldName == column {
				return v.Field(i).Interface()
			}
		}
	}

	return ""
}

// formatRow formats a single row for table output
func (f *TableFormatter) formatRow(data interface{}, columns []string) []string {
	row := make([]string, len(columns))
	for i, col := range columns {
		value := f.getValue(data, col)
		row[i] = fmt.Sprintf("%v", value)
	}
	return row
}
