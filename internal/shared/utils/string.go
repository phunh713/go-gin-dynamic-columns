package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

func BuildFormulaSQL(formula string, contextObj map[string]interface{}) string {
	// in formula, dynamic values are in format {table.column}
	// e.g., {invoices.created_at}, {invoices.payment_terms}
	// This function replaces them with actual values from contextObj
	// contextObj samples map[tableName]tableModel:
	// {
	//   "invoices": {
	//	 	"created_at": time.Time{},
	//	 	"payment_terms": int,
	//   },
	//   "companies": {
	//	 	"status": string,
	//	 	"is_working": bool,
	//   },
	// }

	formula = normalizeFormulaString(formula)

	// Regex to match {table.field} or {table:modifier.field}
	re := regexp.MustCompile(`\{(\w+)(?::(\w+))?\.(\w+)\}`)

	// Find all matches
	matches := re.FindAllStringSubmatch(formula, -1)

	result := formula
	for _, match := range matches {
		placeholder := match[0] // Full match: {invoices:original.created_at}
		tableName := match[1]   // Capture group 1: invoices
		modifier := match[2]    // Capture group 2: original (optional)
		fieldName := match[3]   // Capture group 3: created_at

		// TODO: Handle modifier if needed (e.g., original vs updated values)
		_ = modifier
		ctxKey := tableName
		if modifier != "" {
			ctxKey = fmt.Sprintf("%s:%s", tableName, modifier)
		}

		// Get field value by field name
		contextKey := contextObj[ctxKey]

		// Handle different context value types
		var value any

		// Try to access as map using reflection
		v := reflect.ValueOf(contextKey)
		if v.Kind() == reflect.Map {
			// Access map key
			mapValue := v.MapIndex(reflect.ValueOf(fieldName))
			if mapValue.IsValid() {
				value = mapValue.Interface()
			}
		} else {
			// Try to find field by JSON tag in struct
			value = FindFieldByJsonTag(contextKey, fieldName)
		}

		value = convertValueToSQL(value)

		result = strings.Replace(result, placeholder, fmt.Sprintf("%v", value), 1)
	}

	return result
}

func convertValueToSQL(value interface{}) interface{} {
	valueType := reflect.TypeOf(value)

	if value == nil {
		return "NULL"
	}

	if valueType.Kind() == reflect.String {
		return fmt.Sprintf("'%v'", value)
	}

	if valueType == reflect.TypeOf(time.Time{}) {
		return fmt.Sprintf("'%v'", parseToSQLDate(value.(time.Time)))
	}

	if valueType == reflect.TypeOf(&time.Time{}) {
		timePtr := value.(*time.Time)
		if timePtr == nil {
			return "NULL"
		}
		return fmt.Sprintf("'%v'", parseToSQLDate(*timePtr))
	}

	// if value is a slice of int, convert to comma-separated string
	if valueType.Kind() == reflect.Slice && valueType.Elem().Kind() == reflect.Int {
		s := reflect.ValueOf(value)
		var strValues []string
		for i := 0; i < s.Len(); i++ {
			strValues = append(strValues, fmt.Sprintf("%v", s.Index(i).Interface()))
		}
		return fmt.Sprintf("(%s)", strings.Join(strValues, ","))
	}

	return value
}

func parseToSQLDate(oldDateStr time.Time) string {
	nt := oldDateStr.UTC().Format("2006-01-02 15:04:05-07")
	return nt
}

func FindValueByFieldName(obj interface{}, fieldName string) interface{} {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}
	return field.Interface()
}

func FindFieldByJsonTag(obj interface{}, jsonTag string) interface{} {
	if obj == nil {
		return nil
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Only works on structs
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == jsonTag {
			return v.Field(i).Interface()
		}
	}
	return nil
}

func normalizeFormulaString(formula string) string {
	// Remove new lines, tabs, and extra spaces
	re := regexp.MustCompile(`\s+`)
	normalized := re.ReplaceAllString(formula, " ")
	return strings.TrimSpace(normalized)
}
