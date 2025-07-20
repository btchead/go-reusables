package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/btchead/go-reusables/config/yaml"
	"github.com/go-playground/validator/v10"
)

// Config provides configuration parsing and validation functionality
type Config[T any] struct {
	validator *validator.Validate
	parser    *yaml.Parser[T]
}

// New creates a new Config instance with default validator
func New[T any]() *Config[T] {
	return &Config[T]{
		validator: validator.New(),
		parser:    yaml.NewParser[T](),
	}
}

// NewWithValidator creates a new Config instance with custom validator
func NewWithValidator[T any](v *validator.Validate) *Config[T] {
	return &Config[T]{
		validator: v,
		parser:    yaml.NewParser[T](),
	}
}

// LoadFromFile loads configuration from a YAML file and applies defaults and validation
func (c *Config[T]) LoadFromFile(filename string, target *T) error {
	// First apply defaults
	if err := c.ApplyDefaults(target); err != nil {
		return fmt.Errorf("failed to apply defaults: %w", err)
	}

	// Load from file if it exists
	if c.parser.FileExists(filename) {
		if err := c.parser.ParseFile(filename, target); err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Validate the final configuration
	if err := c.Validate(target); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// LoadFromYAML loads configuration from YAML data and applies defaults and validation
func (c *Config[T]) LoadFromYAML(data []byte, target *T) error {
	// First apply defaults
	if err := c.ApplyDefaults(target); err != nil {
		return fmt.Errorf("failed to apply defaults: %w", err)
	}

	// Parse YAML
	if err := c.parser.Parse(data, target); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate the final configuration
	if err := c.Validate(target); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// ApplyDefaults applies default values from struct tags to the target
func (c *Config[T]) ApplyDefaults(target *T) error {
	return c.applyDefaults(reflect.ValueOf(target))
}

// Validate validates the configuration using the validator package
func (c *Config[T]) Validate(target *T) error {
	return c.validator.Struct(target)
}

// SaveToFile saves the configuration to a YAML file
func (c *Config[T]) SaveToFile(filename string, source *T) error {
	return c.parser.WriteFile(filename, source)
}

// GenerateTemplate creates a YAML template with comments showing defaults and validation rules
func (c *Config[T]) GenerateTemplate() ([]byte, error) {
	generator := yaml.NewGenerator[T]()
	return generator.GenerateTemplate()
}

// GenerateTemplateToFile creates a YAML template file with comments
func (c *Config[T]) GenerateTemplateToFile(filename string) error {
	generator := yaml.NewGenerator[T]()
	return generator.GenerateTemplateToFile(filename)
}

// Load is a convenience function that creates a new config, applies defaults,
// loads from file (if exists), and validates in one call
func Load[T any](filename string) (*T, error) {
	cfg := New[T]()
	var target T

	if err := cfg.LoadFromFile(filename, &target); err != nil {
		return nil, err
	}

	return &target, nil
}

// LoadWithDefaults is a convenience function that applies defaults to the provided target
// then loads from file (if exists) and validates
func LoadWithDefaults[T any](filename string, target *T) error {
	cfg := New[T]()
	return cfg.LoadFromFile(filename, target)
}

// GenerateTemplate creates a YAML template for the specified type
func GenerateTemplate[T any]() ([]byte, error) {
	return yaml.GenerateTemplate[T]()
}

// GenerateTemplateToFile creates a YAML template file for the specified type
func GenerateTemplateToFile[T any](filename string) error {
	return yaml.GenerateTemplateToFile[T](filename)
}

// applyDefaults recursively applies default values
func (c *Config[T]) applyDefaults(v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Handle nested structs
		if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct) {
			if err := c.applyDefaults(field); err != nil {
				return err
			}
			continue
		}

		// Apply default if field is zero value and default tag exists
		defaultValue := fieldType.Tag.Get("default")
		if defaultValue != "" && c.isZeroValue(field) {
			if err := c.setFieldValue(field, defaultValue); err != nil {
				return fmt.Errorf("failed to set default for field %s: %w", fieldType.Name, err)
			}
		}
	}

	return nil
}

// isZeroValue checks if a field contains the zero value for its type
func (c *Config[T]) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map, reflect.Chan:
		return v.IsNil() || v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

// setFieldValue sets a field value from a string representation
func (c *Config[T]) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			// Handle time.Duration
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration: %w", err)
			}
			field.SetInt(int64(duration))
		} else {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer: %w", err)
			}
			field.SetInt(intVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer: %w", err)
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float: %w", err)
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean: %w", err)
		}
		field.SetBool(boolVal)
	case reflect.Slice:
		// Handle slices of strings
		if field.Type().Elem().Kind() == reflect.String {
			values := strings.Split(value, ",")
			for i, v := range values {
				values[i] = strings.TrimSpace(v)
			}
			field.Set(reflect.ValueOf(values))
		} else {
			return fmt.Errorf("unsupported slice type: %s", field.Type())
		}
	default:
		return fmt.Errorf("unsupported field type: %s", field.Type())
	}
	return nil
}
