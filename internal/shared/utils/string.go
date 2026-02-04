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
		value := FindFieldByJsonTag(contextObj[ctxKey], fieldName)
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
