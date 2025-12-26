// Package config provides functionality for parsing krakenv configuration.
package config

import (
	"strings"
)

// KrakenvConfig represents project-level configuration extracted from distributable.
// Configuration is stored as special comments: #krakenv:KEY=VALUE.
type KrakenvConfig struct {
	Environments []string // e.g., ["local", "testing", "production"]
	Strict       bool     // If true, unannotated variables are errors
	DistPath     string   // Override default .env.dist path
}

// DefaultConfig returns a KrakenvConfig with default values.
func DefaultConfig() *KrakenvConfig {
	return &KrakenvConfig{
		Environments: []string{"local"},
		Strict:       false,
		DistPath:     ".env.dist",
	}
}

// ParseConfigLine parses a single #krakenv: configuration line.
// Returns the key and value if valid, or empty strings if not a config line.
func ParseConfigLine(line string) (key, value string) {
	line = strings.TrimSpace(line)

	// Must start with #krakenv:
	if !strings.HasPrefix(line, "#krakenv:") {
		return "", ""
	}

	// Extract key=value part
	content := strings.TrimPrefix(line, "#krakenv:")
	parts := strings.SplitN(content, "=", 2)
	if len(parts) != 2 {
		return "", ""
	}

	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

// IsConfigLine checks if a line is a krakenv configuration line.
func IsConfigLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#krakenv:")
}

// ParseConfig parses all configuration lines and returns a KrakenvConfig.
func ParseConfig(lines []string) *KrakenvConfig {
	config := DefaultConfig()

	for _, line := range lines {
		key, value := ParseConfigLine(line)
		if key == "" {
			continue
		}

		switch key {
		case "environments":
			envs := strings.Split(value, ",")
			config.Environments = make([]string, 0, len(envs))
			for _, env := range envs {
				env = strings.TrimSpace(env)
				if env != "" {
					config.Environments = append(config.Environments, env)
				}
			}
		case "strict":
			config.Strict = value == "true" || value == "1" || value == "yes"
		case "distPath":
			if value != "" {
				config.DistPath = value
			}
		}
	}

	return config
}

// FormatConfigLine formats a configuration key-value pair as a comment line.
func FormatConfigLine(key, value string) string {
	return "#krakenv:" + key + "=" + value
}

// FormatConfig formats a KrakenvConfig as a slice of comment lines.
func FormatConfig(config *KrakenvConfig) []string {
	lines := make([]string, 0, 3)

	if len(config.Environments) > 0 {
		lines = append(lines, FormatConfigLine("environments", strings.Join(config.Environments, ",")))
	}

	if config.Strict {
		lines = append(lines, FormatConfigLine("strict", "true"))
	}

	return lines
}
