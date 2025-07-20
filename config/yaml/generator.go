package yaml

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"
)

// Generator creates YAML templates from struct definitions
type Generator[T any] struct{}

// NewGenerator creates a new YAML template generator
func NewGenerator[T any]() *Generator[T] {
	return &Generator[T]{}
}

// GenerateTemplate creates a YAML template with comments showing default values and validation rules
func (g *Generator[T]) GenerateTemplate() ([]byte, error) {
	var target T
	return g.generateFromStruct(reflect.TypeOf(target), 0)
}

// GenerateTemplateToFile creates a YAML template file with comments
func (g *Generator[T]) GenerateTemplateToFile(filename string) error {
	data, err := g.GenerateTemplate()
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}

// generateFromStruct recursively generates YAML template from struct type
func (g *Generator[T]) generateFromStruct(t reflect.Type, indent int) ([]byte, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type must be a struct, got %s", t.Kind())
	}

	var lines []string
	indentStr := strings.Repeat("  ", indent)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "-" {
			continue
		}

		// Get field name from yaml tag or use field name
		fieldName := field.Name
		if yamlTag != "" {
			parts := strings.Split(yamlTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		// Handle nested structs
		if field.Type.Kind() == reflect.Struct || (field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
			lines = append(lines, indentStr+fieldName+":")

			fieldType := field.Type
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}

			nestedData, err := g.generateFromStructType(fieldType, indent+1)
			if err != nil {
				return nil, err
			}
			lines = append(lines, string(nestedData))
		} else if field.Type.Kind() == reflect.Map && field.Type.Elem().Kind() == reflect.Struct {
			// Handle map[string]StructType
			lines = append(lines, indentStr+fieldName+":")
			
			// Generate a meaningful example key based on struct type name
			structTypeName := field.Type.Elem().Name()
			exampleKey := g.generateExampleKey(structTypeName)
			lines = append(lines, indentStr+"  "+exampleKey+":")
			
			nestedData, err := g.generateFromStructType(field.Type.Elem(), indent+2)
			if err != nil {
				return nil, err
			}
			lines = append(lines, string(nestedData))
		} else {
			// Generate example value
			exampleValue := g.generateExampleValue(field)
			lines = append(lines, indentStr+fieldName+": "+exampleValue)
		}
	}

	return []byte(strings.Join(lines, "\n")), nil
}

// generateFromStructType generates YAML from a specific struct type
func (g *Generator[T]) generateFromStructType(t reflect.Type, indent int) ([]byte, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var lines []string
	indentStr := strings.Repeat("  ", indent)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "-" {
			continue
		}

		fieldName := field.Name
		if yamlTag != "" {
			parts := strings.Split(yamlTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		if field.Type.Kind() == reflect.Struct || (field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
			lines = append(lines, indentStr+fieldName+":")

			fieldType := field.Type
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}

			nestedData, err := g.generateFromStructType(fieldType, indent+1)
			if err != nil {
				return nil, err
			}
			lines = append(lines, string(nestedData))
		} else if field.Type.Kind() == reflect.Map && field.Type.Elem().Kind() == reflect.Struct {
			// Handle map[string]StructType
			lines = append(lines, indentStr+fieldName+":")
			
			// Generate a meaningful example key based on struct type name
			structTypeName := field.Type.Elem().Name()
			exampleKey := g.generateExampleKey(structTypeName)
			lines = append(lines, indentStr+"  "+exampleKey+":")
			
			nestedData, err := g.generateFromStructType(field.Type.Elem(), indent+2)
			if err != nil {
				return nil, err
			}
			lines = append(lines, string(nestedData))
		} else {
			exampleValue := g.generateExampleValue(field)
			lines = append(lines, indentStr+fieldName+": "+exampleValue)
		}
	}

	return []byte(strings.Join(lines, "\n")), nil
}

// generateFieldComment creates a comment describing the field
func (g *Generator[T]) generateFieldComment(field reflect.StructField) string {
	var parts []string

	// Add type information
	parts = append(parts, fmt.Sprintf("Type: %s", g.getTypeDescription(field.Type)))

	// Add default value if present
	if defaultValue := field.Tag.Get("default"); defaultValue != "" {
		parts = append(parts, fmt.Sprintf("Default: %s", defaultValue))
	}

	// Add validation rules if present
	if validate := field.Tag.Get("validate"); validate != "" {
		parts = append(parts, fmt.Sprintf("Validation: %s", validate))
	}

	return strings.Join(parts, " | ")
}

// generateExampleKey creates a meaningful example key for map fields
func (g *Generator[T]) generateExampleKey(structTypeName string) string {
	if structTypeName == "" {
		return "example_key"
	}
	
	// Convert CamelCase to snake_case and add example prefix
	var result strings.Builder
	for i, r := range structTypeName {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(rune(strings.ToLower(string(r))[0]))
	}
	
	return "example_" + result.String()
}

// generateExampleValue creates an example value for a field
func (g *Generator[T]) generateExampleValue(field reflect.StructField) string {
	// Use default value if available
	if defaultValue := field.Tag.Get("default"); defaultValue != "" {
		return g.formatExampleValue(field.Type, defaultValue)
	}

	// Generate type-appropriate example
	return g.generateTypeExample(field.Type)
}

// formatExampleValue formats a default value appropriately for YAML
func (g *Generator[T]) formatExampleValue(fieldType reflect.Type, value string) string {
	switch fieldType.Kind() {
	case reflect.String:
		return fmt.Sprintf(`"%s"`, value)
	case reflect.Slice:
		if fieldType.Elem().Kind() == reflect.String {
			// Handle comma-separated values
			if strings.Contains(value, ",") {
				items := strings.Split(value, ",")
				var yamlItems []string
				for _, item := range items {
					yamlItems = append(yamlItems, fmt.Sprintf(`  - "%s"`, strings.TrimSpace(item)))
				}
				return "\n" + strings.Join(yamlItems, "\n")
			}
		}
		return value
	default:
		return value
	}
}

// generateTypeExample generates example values based on type
func (g *Generator[T]) generateTypeExample(fieldType reflect.Type) string {
	if fieldType == reflect.TypeOf(time.Duration(0)) {
		return `"30s"`
	}

	switch fieldType.Kind() {
	case reflect.String:
		return `"example_value"`
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "0"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "0"
	case reflect.Float32, reflect.Float64:
		return "0.0"
	case reflect.Bool:
		return "false"
	case reflect.Slice:
		if fieldType.Elem().Kind() == reflect.String {
			return "\n  - \"item1\"\n  - \"item2\""
		}
		return "[]"
	case reflect.Map:
		return "{}"
	case reflect.Ptr:
		return g.generateTypeExample(fieldType.Elem())
	default:
		return `"TODO: configure this field"`
	}
}

// getTypeDescription returns a human-readable type description
func (g *Generator[T]) getTypeDescription(fieldType reflect.Type) string {
	if fieldType == reflect.TypeOf(time.Duration(0)) {
		return "duration (e.g., '30s', '5m', '1h')"
	}

	switch fieldType.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "unsigned integer"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice:
		elemDesc := g.getTypeDescription(fieldType.Elem())
		return fmt.Sprintf("array of %s", elemDesc)
	case reflect.Map:
		return "map"
	case reflect.Ptr:
		return g.getTypeDescription(fieldType.Elem())
	case reflect.Struct:
		return "object"
	default:
		return fieldType.String()
	}
}

// Convenience functions

// GenerateTemplate creates a YAML template for the specified type
func GenerateTemplate[T any]() ([]byte, error) {
	generator := NewGenerator[T]()
	return generator.GenerateTemplate()
}

// GenerateTemplateToFile creates a YAML template file for the specified type
func GenerateTemplateToFile[T any](filename string) error {
	generator := NewGenerator[T]()
	return generator.GenerateTemplateToFile(filename)
}
