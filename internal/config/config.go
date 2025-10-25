package core

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

// Resource defines a file to be downloaded
type Resource struct {
	URL  string `yaml:"url"`
	Path string `yaml:"path"`
}

// Config is the main project configuration
type Config struct {
	ProjectName  string              `yaml:"project_name"`
	Sources      []string            `yaml:"sources,omitempty"`
	Output       string              `yaml:"output,omitempty"`
	Flags        []string            `yaml:"flags,omitempty"`
	Dependencies map[string][]string `yaml:"dependencies,omitempty"`
	Resources    []Resource          `yaml:"resources,omitempty"`
	// Optional stuff to add
	Author      string                    `yaml:"author,omitempty"`
	Description string                    `yaml:"description,omitempty"`
	Env         map[string]string         `yaml:"env,omitempty"`
	Platforms   map[string]PlatformConfig `yaml:"platforms,omitempty"`
	CreatedAt   string                    `yaml:"created_at,omitempty"`
}

// PlatformConfig allows OS-specific overrides for dependencies or resources
type PlatformConfig struct {
	Dependencies []string   `yaml:"dependencies,omitempty"`
	Resources    []Resource `yaml:"resources,omitempty"`
}

// LoadConfig reads and parses a YAML configuration file into Config
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML syntax: %w", err)
	}

	// Fill missing metadata dynamically
	if cfg.CreatedAt == "" {
		cfg.CreatedAt = time.Now().Format(time.RFC3339)
	}

	return &cfg, nil
}

// GetDependencies returns the dependency list for the current OS
func (c *Config) GetDependencies() []string {
	osKey := runtime.GOOS

	// 1. OS-specific overrides
	if platform, ok := c.Platforms[osKey]; ok && len(platform.Dependencies) > 0 {
		return platform.Dependencies
	}

	// 2. Top-level dependency map
	if deps, ok := c.Dependencies[osKey]; ok {
		return deps
	}

	// 3. Backward compatibility: check for "macos" if osKey is "darwin"
	if osKey == "darwin" {
		if deps, ok := c.Dependencies["macos"]; ok {
			return deps
		}
	}

	// 4. Default fallback
	return []string{}
}
