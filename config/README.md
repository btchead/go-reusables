# 🔧 Configuration Package

A powerful configuration management package for Go applications that supports struct tag defaults, YAML file parsing, and validation using the validator package.

## Features

- **Struct Tag Defaults**: Automatically apply default values using `default` struct tags
- **YAML Support**: Load configuration from YAML files or raw YAML data
- **Validation**: Integrate with `github.com/go-playground/validator/v10` for comprehensive validation
- **Type Safety**: Support for various Go types including `time.Duration`, slices, and nested structs
- **Zero Configuration**: Works out of the box with sensible defaults
- **File Operations**: Save configurations back to YAML files

## Installation

```bash
go get github.com/btchead/go-reusables/config
```

## Quick Start

### Define Configuration Struct

```go
package main

import (
    "time"
    "github.com/btchead/go-reusables/config"
)

type DatabaseConfig struct {
    Host     string        `yaml:"host" default:"localhost" validate:"required"`
    Port     int           `yaml:"port" default:"5432" validate:"min=1,max=65535"`
    Username string        `yaml:"username" default:"postgres" validate:"required"`
    Password string        `yaml:"password" validate:"required"`
    Database string        `yaml:"database" default:"myapp" validate:"required"`
    Timeout  time.Duration `yaml:"timeout" default:"30s" validate:"required"`
    MaxConns int           `yaml:"max_connections" default:"10" validate:"min=1"`
}

type ServerConfig struct {
    Host string `yaml:"host" default:"0.0.0.0" validate:"required"`
    Port int    `yaml:"port" default:"8080" validate:"min=1,max=65535"`
}

type AppConfig struct {
    Database DatabaseConfig `yaml:"database"`
    Server   ServerConfig   `yaml:"server"`
    Debug    bool           `yaml:"debug" default:"false"`
    Features []string       `yaml:"features" default:"feature1,feature2"`
}
```

### Load Configuration

```go
func main() {
    // Method 1: Using the generic API
    cfg := config.New[AppConfig]()
    
    var appConfig AppConfig
    err := cfg.LoadFromFile("config.yaml", &appConfig)
    if err != nil {
        log.Fatal(err)
    }
    
    // Method 2: Using convenience function
    appConfig2, err := config.Load[AppConfig]("config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Database: %s:%d\n", appConfig.Database.Host, appConfig.Database.Port)
    fmt.Printf("Server: %s:%d\n", appConfig.Server.Host, appConfig.Server.Port)
}
```

## Supported Features

### Default Values

Use the `default` struct tag to specify default values:

```go
type Config struct {
    Host     string        `default:"localhost"`
    Port     int           `default:"8080"`
    Enabled  bool          `default:"true"`
    Timeout  time.Duration `default:"30s"`
    Features []string      `default:"auth,logging,metrics"`
}
```

### Supported Types

- **Basic types**: `string`, `int`, `uint`, `float`, `bool`
- **Time duration**: `time.Duration` (e.g., "30s", "5m", "1h")
- **Slices**: `[]string` (comma-separated values)
- **Nested structs**: Recursively applies defaults and validation

### Validation

Use standard validator tags for validation:

```go
type Config struct {
    Email    string `validate:"required,email"`
    Age      int    `validate:"min=18,max=120"`
    Username string `validate:"required,alphanum,min=3,max=20"`
    Website  string `validate:"url"`
}
```

### YAML Configuration

Create a `config.yaml` file:

```yaml
database:
  host: prod-db.example.com
  port: 5432
  username: produser
  password: secretpassword
  database: production
  timeout: 60s
  max_connections: 20

server:
  host: 0.0.0.0
  port: 3000

debug: false
features:
  - authentication
  - logging
  - metrics
  - monitoring
```

## API Reference

### Config Methods

```go
// Create new config instance with type safety
cfg := config.New[AppConfig]()

// Create with custom validator
cfg := config.NewWithValidator[AppConfig](validator.New())

// Load from file
var appConfig AppConfig
err := cfg.LoadFromFile("config.yaml", &appConfig)

// Load from YAML data
err := cfg.LoadFromYAML(yamlData, &appConfig)

// Apply defaults only
err := cfg.ApplyDefaults(&appConfig)

// Validate configuration
err := cfg.Validate(&appConfig)

// Save to file
err := cfg.SaveToFile("config.yaml", &appConfig)

// Convenience functions
appConfig, err := config.Load[AppConfig]("config.yaml")
err = config.LoadWithDefaults("config.yaml", &appConfig)
```

## Advanced Usage

### Custom Validator

```go
import "github.com/go-playground/validator/v10"

// Create custom validator with custom validation rules
validate := validator.New()

// Register custom validation
validate.RegisterValidation("custom", func(fl validator.FieldLevel) bool {
    return fl.Field().String() != "forbidden"
})

cfg := config.NewWithValidator[AppConfig](validate)
```

### Programmatic Configuration

```go
cfg := config.New[AppConfig]()

// Apply defaults without loading from file
var appConfig AppConfig
err := cfg.ApplyDefaults(&appConfig)

// Manually set values
appConfig.Database.Host = "custom-host"

// Validate the configuration
err = cfg.Validate(&appConfig)

// Save to file
err = cfg.SaveToFile("generated-config.yaml", &appConfig)
```

### Error Handling

```go
err := cfg.LoadFromFile("config.yaml", &appConfig)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "validation failed"):
        log.Printf("Configuration validation error: %v", err)
    case strings.Contains(err.Error(), "failed to read config file"):
        log.Printf("File reading error: %v", err)
    case strings.Contains(err.Error(), "failed to parse YAML"):
        log.Printf("YAML parsing error: %v", err)
    default:
        log.Printf("Configuration error: %v", err)
    }
}
```

## Best Practices

1. **Use Validation**: Always validate your configuration to catch errors early
2. **Provide Defaults**: Use default tags for optional configuration values
3. **Document Requirements**: Use validation tags to document field requirements
4. **Nested Structs**: Organize related configuration into nested structs
5. **Environment Specific**: Use different config files for different environments

## License

MIT License - see [LICENSE](../LICENSE) file for details.