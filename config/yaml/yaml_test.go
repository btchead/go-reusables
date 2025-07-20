package yaml

import (
	"os"
	"strings"
	"testing"
	"time"
)

type TestConfig struct {
	StringField   string        `yaml:"string_field" default:"default_string" validate:"required"`
	IntField      int           `yaml:"int_field" default:"42" validate:"min=1"`
	BoolField     bool          `yaml:"bool_field" default:"true"`
	DurationField time.Duration `yaml:"duration_field" default:"5m" validate:"required"`
	SliceField    []string      `yaml:"slice_field" default:"item1,item2,item3"`
	NestedField   NestedConfig  `yaml:"nested"`
}

type NestedConfig struct {
	NestedString string `yaml:"nested_string" default:"nested_default" validate:"required"`
	NestedInt    int    `yaml:"nested_int" default:"100" validate:"min=50"`
}

func TestParser_Parse(t *testing.T) {
	parser := NewParser[TestConfig]()

	yamlData := []byte(`
string_field: "test_string"
int_field: 123
bool_field: false
duration_field: "10m"
slice_field: ["test1", "test2"]
nested:
  nested_string: "test_nested"
  nested_int: 200
`)

	var config TestConfig
	err := parser.Parse(yamlData, &config)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if config.StringField != "test_string" {
		t.Errorf("Expected StringField 'test_string', got '%s'", config.StringField)
	}

	if config.IntField != 123 {
		t.Errorf("Expected IntField 123, got %d", config.IntField)
	}

	if config.BoolField != false {
		t.Errorf("Expected BoolField false, got %t", config.BoolField)
	}

	if config.DurationField != 10*time.Minute {
		t.Errorf("Expected DurationField 10m, got %v", config.DurationField)
	}

	if config.NestedField.NestedString != "test_nested" {
		t.Errorf("Expected NestedString 'test_nested', got '%s'", config.NestedField.NestedString)
	}
}

func TestParser_ParseFile(t *testing.T) {
	parser := NewParser[TestConfig]()

	yamlContent := `
string_field: "file_string"
int_field: 456
bool_field: true
duration_field: "15m"
slice_field: ["file1", "file2", "file3"]
nested:
  nested_string: "file_nested"
  nested_int: 300
`

	tempFile := "test_parse.yaml"
	err := os.WriteFile(tempFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tempFile)

	var config TestConfig
	err = parser.ParseFile(tempFile, &config)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if config.StringField != "file_string" {
		t.Errorf("Expected StringField 'file_string', got '%s'", config.StringField)
	}

	if config.IntField != 456 {
		t.Errorf("Expected IntField 456, got %d", config.IntField)
	}

	if config.NestedField.NestedInt != 300 {
		t.Errorf("Expected NestedInt 300, got %d", config.NestedField.NestedInt)
	}
}

func TestParser_Marshal(t *testing.T) {
	parser := NewParser[TestConfig]()

	config := TestConfig{
		StringField:   "marshal_test",
		IntField:      789,
		BoolField:     true,
		DurationField: 20 * time.Minute,
		SliceField:    []string{"marshal1", "marshal2"},
		NestedField: NestedConfig{
			NestedString: "marshal_nested",
			NestedInt:    400,
		},
	}

	data, err := parser.Marshal(&config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	yamlString := string(data)
	if !strings.Contains(yamlString, "marshal_test") {
		t.Errorf("YAML should contain 'marshal_test'")
	}

	if !strings.Contains(yamlString, "789") {
		t.Errorf("YAML should contain '789'")
	}

	if !strings.Contains(yamlString, "marshal_nested") {
		t.Errorf("YAML should contain 'marshal_nested'")
	}
}

func TestParser_WriteFile(t *testing.T) {
	parser := NewParser[TestConfig]()

	config := TestConfig{
		StringField:   "write_test",
		IntField:      999,
		BoolField:     false,
		DurationField: 30 * time.Minute,
		SliceField:    []string{"write1", "write2"},
		NestedField: NestedConfig{
			NestedString: "write_nested",
			NestedInt:    500,
		},
	}

	tempFile := "test_write.yaml"
	defer os.Remove(tempFile)

	err := parser.WriteFile(tempFile, &config)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify file was created
	if !parser.FileExists(tempFile) {
		t.Error("File should exist after WriteFile")
	}

	// Read back and verify content
	var readConfig TestConfig
	err = parser.ParseFile(tempFile, &readConfig)
	if err != nil {
		t.Fatalf("Failed to read back written file: %v", err)
	}

	if readConfig.StringField != config.StringField {
		t.Errorf("Read config mismatch: expected '%s', got '%s'", config.StringField, readConfig.StringField)
	}
}

func TestParser_FileExists(t *testing.T) {
	parser := NewParser[TestConfig]()

	// Test non-existent file
	if parser.FileExists("nonexistent.yaml") {
		t.Error("FileExists should return false for non-existent file")
	}

	// Create a file and test
	tempFile := "test_exists.yaml"
	err := os.WriteFile(tempFile, []byte("test: true"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tempFile)

	if !parser.FileExists(tempFile) {
		t.Error("FileExists should return true for existing file")
	}
}

func TestConvenienceFunctions(t *testing.T) {
	t.Run("Parse function", func(t *testing.T) {
		yamlData := []byte(`
string_field: "convenience_test"
int_field: 111
`)

		var config TestConfig
		err := Parse(yamlData, &config)
		if err != nil {
			t.Fatalf("Parse convenience function failed: %v", err)
		}

		if config.StringField != "convenience_test" {
			t.Errorf("Expected 'convenience_test', got '%s'", config.StringField)
		}
	})

	t.Run("WriteFile function", func(t *testing.T) {
		config := TestConfig{
			StringField: "convenience_write",
			IntField:    222,
		}

		tempFile := "test_convenience.yaml"
		defer os.Remove(tempFile)

		err := WriteFile(tempFile, &config)
		if err != nil {
			t.Fatalf("WriteFile convenience function failed: %v", err)
		}

		// Verify content
		var readConfig TestConfig
		err = ParseFile(tempFile, &readConfig)
		if err != nil {
			t.Fatalf("Failed to read convenience written file: %v", err)
		}

		if readConfig.StringField != "convenience_write" {
			t.Errorf("Expected 'convenience_write', got '%s'", readConfig.StringField)
		}
	})
}

func TestGenerator_GenerateTemplate(t *testing.T) {
	generator := NewGenerator[TestConfig]()

	template, err := generator.GenerateTemplate()
	if err != nil {
		t.Fatalf("GenerateTemplate failed: %v", err)
	}

	templateString := string(template)

	// Check that template contains field names
	if !strings.Contains(templateString, "string_field:") {
		t.Error("Template should contain 'string_field:'")
	}

	if !strings.Contains(templateString, "int_field:") {
		t.Error("Template should contain 'int_field:'")
	}

	if !strings.Contains(templateString, "nested:") {
		t.Error("Template should contain 'nested:'")
	}

	// Check default values are used
	if !strings.Contains(templateString, "default_string") {
		t.Error("Template should contain default string value")
	}

	if !strings.Contains(templateString, "42") {
		t.Error("Template should contain default int value")
	}
}

func TestGenerator_GenerateTemplateToFile(t *testing.T) {
	generator := NewGenerator[TestConfig]()

	tempFile := "test_template.yaml"
	defer os.Remove(tempFile)

	err := generator.GenerateTemplateToFile(tempFile)
	if err != nil {
		t.Fatalf("GenerateTemplateToFile failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("Template file should have been created")
	}

	// Read and verify content
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read template file: %v", err)
	}

	contentString := string(content)
	if !strings.Contains(contentString, "string_field:") {
		t.Error("Template file should contain field definitions")
	}
}

func TestGeneratorConvenienceFunctions(t *testing.T) {
	t.Run("GenerateTemplate function", func(t *testing.T) {
		template, err := GenerateTemplate[TestConfig]()
		if err != nil {
			t.Fatalf("GenerateTemplate convenience function failed: %v", err)
		}

		templateString := string(template)
		if !strings.Contains(templateString, "string_field:") {
			t.Error("Template should contain field definitions")
		}
	})

	t.Run("GenerateTemplateToFile function", func(t *testing.T) {
		tempFile := "test_convenience_template.yaml"
		defer os.Remove(tempFile)

		err := GenerateTemplateToFile[TestConfig](tempFile)
		if err != nil {
			t.Fatalf("GenerateTemplateToFile convenience function failed: %v", err)
		}

		if _, err := os.Stat(tempFile); os.IsNotExist(err) {
			t.Error("Template file should have been created")
		}
	})
}
