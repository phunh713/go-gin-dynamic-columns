package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

func PrettyPrintRoutes(routes gin.RoutesInfo) {
	// ANSI color codes
	const (
		colorReset  = "\033[0m"
		colorRed    = "\033[31m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorCyan   = "\033[36m"
		colorBold   = "\033[1m"
	)

	fmt.Printf("\n%s%s=== Registered Routes ===%s\n", colorBold, colorCyan, colorReset)
	for _, route := range routes {
		// Color code based on HTTP method
		var methodColor string
		switch route.Method {
		case "GET":
			methodColor = colorGreen
		case "POST":
			methodColor = colorYellow
		case "PUT":
			methodColor = colorBlue
		case "DELETE":
			methodColor = colorRed
		default:
			methodColor = colorCyan
		}

		fmt.Printf("%s%-8s%s %s%-30s%s %s-->%s %s\n",
			methodColor, route.Method, colorReset,
			colorCyan, route.Path, colorReset,
			colorBold, colorReset,
			route.Handler)
	}
	fmt.Printf("%s%s=========================%s\n", colorBold, colorCyan, colorReset)
}

// NewInstance creates a new addressable instance from a model template
// If modelTemplate is a pointer, returns a new pointer to the same type
// If modelTemplate is a struct, returns a pointer to a new struct instance
func NewInstance(modelTemplate any) any {
	modelType := reflect.TypeOf(modelTemplate)

	// If it's a pointer, get the element type
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Create new instance and return pointer to it
	newValue := reflect.New(modelType)
	return newValue.Interface()
}

func GetStructFieldNames(model interface{}) []string {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var fieldNames []string
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldNames = append(fieldNames, field.Name)
	}
	return fieldNames
}

func GetStructFieldJsonTags(model interface{}) []string {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var jsonTags []string
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag != "" && tag != "-" {
			// Extract field name before comma (removes omitempty, string, etc.)
			fieldName := strings.Split(tag, ",")[0]
			jsonTags = append(jsonTags, fieldName)
		}
	}
	return jsonTags
}

// GetNonZeroStructFieldJsonTags returns only JSON tags of fields with non-zero values
// Useful for tracking which fields were actually set in an UPDATE request
func GetNonZeroStructFieldJsonTags(model interface{}) []string {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var jsonTags []string
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		tag := field.Tag.Get("json")
		if tag != "" && tag != "-" {
			// Extract field name before comma (removes omitempty, string, etc.)
			fieldName := strings.Split(tag, ",")[0]

			// Only include fields that have non-zero values (were actually set in JSON)
			if !fieldValue.IsZero() {
				jsonTags = append(jsonTags, fieldName)
			}
		}
	}
	return jsonTags
}
