package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	// TasksPath is the path to the tasks directory.
	TasksPath string `yaml:"tasks_path"`

	// LogLevel is the logging level: debug, info, warn, error.
	LogLevel string `yaml:"log_level"`

	// Server contains MCP server-specific configuration.
	Server ServerConfig `yaml:"server"`
}

// ServerConfig contains MCP server configuration.
type ServerConfig struct {
	// Name is the server name exposed via MCP.
	Name string `yaml:"name"`

	// Version is the server version.
	Version string `yaml:"version"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		TasksPath: "",
		LogLevel:  "info",
		Server: ServerConfig{
			Name:    "agent-memory",
			Version: "1.0.0",
		},
	}
}

// Load reads configuration from a YAML file.
// If the file doesn't exist, returns default configuration.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// LoadFromDefaultLocations tries to load config from standard locations:
// 1. ./agent-memory.yaml (current directory)
// 2. ~/.agent-memory/config.yaml (home directory)
// 3. /etc/agent-memory/config.yaml (system-wide, Unix only)
// Returns default config if no file found.
func LoadFromDefaultLocations() (*Config, error) {
	locations := []string{
		"agent-memory.yaml",
	}

	// Add home directory config
	if homeDir, err := os.UserHomeDir(); err == nil {
		locations = append(locations, filepath.Join(homeDir, ".agent-memory", "config.yaml"))
	}

	// Add system-wide config (Unix)
	locations = append(locations, "/etc/agent-memory/config.yaml")

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return Load(loc)
		}
	}

	return DefaultConfig(), nil
}

// ResolveTasksPath returns the resolved tasks path.
// If TasksPath is empty, returns ~/.agent-memory/tasks.
func (c *Config) ResolveTasksPath() (string, error) {
	if c.TasksPath != "" {
		// Expand ~ if present
		if len(c.TasksPath) > 0 && c.TasksPath[0] == '~' {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get home directory: %w", err)
			}
			return filepath.Join(homeDir, c.TasksPath[1:]), nil
		}
		return c.TasksPath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".agent-memory", "tasks"), nil
}

// Save writes the configuration to a YAML file.
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
