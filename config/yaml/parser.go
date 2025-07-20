package yaml

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Parser handles YAML parsing operations
type Parser[T any] struct{}

// NewParser creates a new YAML parser for the specified type
func NewParser[T any]() *Parser[T] {
	return &Parser[T]{}
}

// ParseFile reads and parses a YAML file into the target struct
func (p *Parser[T]) ParseFile(filename string, target *T) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	return p.Parse(data, target)
}

// Parse parses YAML data into the target struct
func (p *Parser[T]) Parse(data []byte, target *T) error {
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}
	return nil
}

// Marshal converts a struct to YAML bytes
func (p *Parser[T]) Marshal(source *T) ([]byte, error) {
	data, err := yaml.Marshal(source)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return data, nil
}

// WriteFile writes a struct to a YAML file
func (p *Parser[T]) WriteFile(filename string, source *T) error {
	data, err := p.Marshal(source)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	return nil
}

// FileExists checks if a YAML file exists
func (p *Parser[T]) FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// Convenience functions for direct usage without creating a parser instance

// ParseFile reads and parses a YAML file into the target struct
func ParseFile[T any](filename string, target *T) error {
	parser := NewParser[T]()
	return parser.ParseFile(filename, target)
}

// Parse parses YAML data into the target struct
func Parse[T any](data []byte, target *T) error {
	parser := NewParser[T]()
	return parser.Parse(data, target)
}

// Marshal converts a struct to YAML bytes
func Marshal[T any](source *T) ([]byte, error) {
	parser := NewParser[T]()
	return parser.Marshal(source)
}

// WriteFile writes a struct to a YAML file
func WriteFile[T any](filename string, source *T) error {
	parser := NewParser[T]()
	return parser.WriteFile(filename, source)
}