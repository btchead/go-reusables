package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

type TestAppConfig struct {
	NestedConfig  TestNestedConfig            `yaml:"nested_config"`
	Server        TestServerConfig            `yaml:"server"`
	Debug         bool                        `yaml:"debug" default:"false"`
	Features      []string                    `yaml:"features" default:"feature1,feature2"`
	NestedConfigs map[string]TestNestedConfig `yaml:"nested_configs"`
}

type TestNestedConfig struct {
	Timeout time.Duration `yaml:"timeout" default:"30s" validate:"required"`
}

type TestServerConfig struct {
	Host string `yaml:"host" default:"0.0.0.0" validate:"required"`
	Port int    `yaml:"port" default:"8080" validate:"min=1,max=65535"`
}

func TestConfig_LoadFromFile_WithYamlParser(t *testing.T) {
	cfg := New[TestAppConfig]()

	var appConfig TestAppConfig
	err := cfg.LoadFromFile("nonexistent.yaml", &appConfig)
	if err != nil {
		t.Fatalf("LoadFromFile should not fail for non-existent file: %v", err)
	}

	if appConfig.Server.Port != 8080 {
		t.Errorf("Expected default server port 8080, got %d", appConfig.Server.Port)
	}
}

func TestConfig_GenerateTemplate(t *testing.T) {
	cfg := New[TestAppConfig]()

	template, err := cfg.GenerateTemplate()
	if err != nil {
		t.Fatalf("GenerateTemplate failed: %v", err)
	}

	templateString := string(template)

	if !strings.Contains(templateString, "server:") {
		t.Error("Template should contain 'server:'")
	}

	if !strings.Contains(templateString, "host:") {
		t.Error("Template should contain 'host:'")
	}
}

func TestConfig_GenerateTemplateToFile(t *testing.T) {
	cfg := New[TestAppConfig]()

	tempFile := "test_template.yaml"
	defer os.Remove(tempFile)

	err := cfg.GenerateTemplateToFile(tempFile)
	if err != nil {
		t.Fatalf("GenerateTemplateToFile failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("Template file should have been created")
	}
}
