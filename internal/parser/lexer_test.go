package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenizeLine_SimpleVariable(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantVal  string
		wantErr  bool
	}{
		{
			name:     "simple variable",
			input:    "DB_HOST=localhost",
			wantName: "DB_HOST",
			wantVal:  "localhost",
		},
		{
			name:     "variable with empty value",
			input:    "DB_HOST=",
			wantName: "DB_HOST",
			wantVal:  "",
		},
		{
			name:     "variable with spaces in value",
			input:    "MESSAGE=hello world",
			wantName: "MESSAGE",
			wantVal:  "hello world",
		},
		{
			name:     "variable with quoted value",
			input:    `DB_HOST="localhost"`,
			wantName: "DB_HOST",
			wantVal:  "localhost",
		},
		{
			name:     "variable with single quoted value",
			input:    `DB_HOST='localhost'`,
			wantName: "DB_HOST",
			wantVal:  "localhost",
		},
		{
			name:     "variable with equals in value",
			input:    "CONN_STRING=host=localhost;port=5432",
			wantName: "CONN_STRING",
			wantVal:  "host=localhost;port=5432",
		},
		{
			name:     "variable with number value",
			input:    "PORT=5432",
			wantName: "PORT",
			wantVal:  "5432",
		},
		{
			name:     "whitespace trimmed",
			input:    "  DB_HOST = localhost  ",
			wantName: "DB_HOST",
			wantVal:  "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, value, _, err := TokenizeLine(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, name)
			assert.Equal(t, tt.wantVal, value)
		})
	}
}

func TestTokenizeLine_WithAnnotation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantName       string
		wantVal        string
		wantAnnotation string
	}{
		{
			name:           "simple annotation",
			input:          "DB_HOST=localhost #prompt:Database host?|string",
			wantName:       "DB_HOST",
			wantVal:        "localhost",
			wantAnnotation: "#prompt:Database host?|string",
		},
		{
			name:           "annotation with constraints",
			input:          "PORT=5432 #prompt:Port?|int;min:1;max:65535",
			wantName:       "PORT",
			wantVal:        "5432",
			wantAnnotation: "#prompt:Port?|int;min:1;max:65535",
		},
		{
			name:           "empty value with annotation",
			input:          "PASSWORD= #prompt:Password?|string;secret",
			wantName:       "PASSWORD",
			wantVal:        "",
			wantAnnotation: "#prompt:Password?|string;secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, value, annotation, err := TokenizeLine(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, name)
			assert.Equal(t, tt.wantVal, value)
			assert.Equal(t, tt.wantAnnotation, annotation)
		})
	}
}

func TestTokenizeLine_Comments(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantEmpty bool
	}{
		{
			name:      "full line comment",
			input:     "# This is a comment",
			wantEmpty: true,
		},
		{
			name:      "empty line",
			input:     "",
			wantEmpty: true,
		},
		{
			name:      "whitespace only",
			input:     "   ",
			wantEmpty: true,
		},
		{
			name:      "krakenv config line",
			input:     "#krakenv:environments=local,prod",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, _, _, err := TokenizeLine(tt.input)
			require.NoError(t, err)
			if tt.wantEmpty {
				assert.Empty(t, name)
			}
		})
	}
}

func TestTokenizeLine_InvalidVariableName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "lowercase not allowed",
			input:   "db_host=localhost",
			wantErr: true,
		},
		{
			name:    "starts with number",
			input:   "1VAR=value",
			wantErr: true,
		},
		{
			name:    "contains special chars",
			input:   "VAR-NAME=value",
			wantErr: true,
		},
		{
			name:    "valid uppercase",
			input:   "VAR_NAME=value",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := TokenizeLine(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
