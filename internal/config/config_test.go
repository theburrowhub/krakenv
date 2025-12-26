package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfigLine(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantKey   string
		wantValue string
	}{
		{
			name:      "environments config",
			input:     "#krakenv:environments=local,testing,production",
			wantKey:   "environments",
			wantValue: "local,testing,production",
		},
		{
			name:      "strict config",
			input:     "#krakenv:strict=true",
			wantKey:   "strict",
			wantValue: "true",
		},
		{
			name:      "distPath config",
			input:     "#krakenv:distPath=config/.env.dist",
			wantKey:   "distPath",
			wantValue: "config/.env.dist",
		},
		{
			name:      "with leading whitespace",
			input:     "  #krakenv:strict=false",
			wantKey:   "strict",
			wantValue: "false",
		},
		{
			name:      "not a config line",
			input:     "# This is a comment",
			wantKey:   "",
			wantValue: "",
		},
		{
			name:      "variable line",
			input:     "DB_HOST=localhost",
			wantKey:   "",
			wantValue: "",
		},
		{
			name:      "empty line",
			input:     "",
			wantKey:   "",
			wantValue: "",
		},
		{
			name:      "missing equals",
			input:     "#krakenv:stricttrue",
			wantKey:   "",
			wantValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value := ParseConfigLine(tt.input)
			assert.Equal(t, tt.wantKey, key)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}

func TestIsConfigLine(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"#krakenv:environments=local", true},
		{"  #krakenv:strict=true", true},
		{"# Regular comment", false},
		{"DB_HOST=localhost", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, IsConfigLine(tt.input))
		})
	}
}

func TestParseConfig(t *testing.T) {
	lines := []string{
		"#krakenv:environments=local,testing,production",
		"#krakenv:strict=true",
		"#krakenv:distPath=config/.env.dist",
	}

	config := ParseConfig(lines)
	require.NotNil(t, config)

	assert.Equal(t, []string{"local", "testing", "production"}, config.Environments)
	assert.True(t, config.Strict)
	assert.Equal(t, "config/.env.dist", config.DistPath)
}

func TestParseConfig_Defaults(t *testing.T) {
	config := ParseConfig(nil)
	require.NotNil(t, config)

	assert.Equal(t, []string{"local"}, config.Environments)
	assert.False(t, config.Strict)
	assert.Equal(t, ".env.dist", config.DistPath)
}

func TestParseConfig_PartialOverride(t *testing.T) {
	lines := []string{
		"#krakenv:environments=dev,prod",
	}

	config := ParseConfig(lines)
	require.NotNil(t, config)

	assert.Equal(t, []string{"dev", "prod"}, config.Environments)
	assert.False(t, config.Strict)                // Default
	assert.Equal(t, ".env.dist", config.DistPath) // Default
}

func TestParseConfig_StrictValues(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"#krakenv:strict=true", true},
		{"#krakenv:strict=1", true},
		{"#krakenv:strict=yes", true},
		{"#krakenv:strict=false", false},
		{"#krakenv:strict=0", false},
		{"#krakenv:strict=no", false},
		{"#krakenv:strict=invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			config := ParseConfig([]string{tt.input})
			assert.Equal(t, tt.want, config.Strict)
		})
	}
}

func TestParseConfig_EnvironmentsTrimmed(t *testing.T) {
	lines := []string{
		"#krakenv:environments= local , testing , production ",
	}

	config := ParseConfig(lines)
	assert.Equal(t, []string{"local", "testing", "production"}, config.Environments)
}

func TestParseConfig_EmptyEnvironments(t *testing.T) {
	lines := []string{
		"#krakenv:environments=",
	}

	config := ParseConfig(lines)
	assert.Empty(t, config.Environments)
}

func TestFormatConfigLine(t *testing.T) {
	line := FormatConfigLine("environments", "local,prod")
	assert.Equal(t, "#krakenv:environments=local,prod", line)
}

func TestFormatConfig(t *testing.T) {
	config := &KrakenvConfig{
		Environments: []string{"local", "prod"},
		Strict:       true,
		DistPath:     ".env.dist", // Default, not included
	}

	lines := FormatConfig(config)
	require.Len(t, lines, 2)
	assert.Equal(t, "#krakenv:environments=local,prod", lines[0])
	assert.Equal(t, "#krakenv:strict=true", lines[1])
}

func TestFormatConfig_NoStrict(t *testing.T) {
	config := &KrakenvConfig{
		Environments: []string{"local"},
		Strict:       false,
	}

	lines := FormatConfig(config)
	require.Len(t, lines, 1)
	assert.Equal(t, "#krakenv:environments=local", lines[0])
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	require.NotNil(t, config)

	assert.Equal(t, []string{"local"}, config.Environments)
	assert.False(t, config.Strict)
	assert.Equal(t, ".env.dist", config.DistPath)
}
