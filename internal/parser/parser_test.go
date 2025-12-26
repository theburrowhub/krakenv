package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAnnotation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantPrompt     string
		wantType       VariableType
		wantOptional   bool
		wantSecret     bool
		wantConstraint map[string]string
		wantErr        bool
	}{
		{
			name:       "simple string",
			input:      "#prompt:Enter value?|string",
			wantPrompt: "Enter value?",
			wantType:   TypeString,
		},
		{
			name:       "int with min max",
			input:      "#prompt:Port?|int;min:1;max:65535",
			wantPrompt: "Port?",
			wantType:   TypeInt,
			wantConstraint: map[string]string{
				"min": "1",
				"max": "65535",
			},
		},
		{
			name:       "enum with options",
			input:      "#prompt:Environment?|enum;options:dev,staging,prod",
			wantPrompt: "Environment?",
			wantType:   TypeEnum,
			wantConstraint: map[string]string{
				"options": "dev,staging,prod",
			},
		},
		{
			name:         "optional modifier",
			input:        "#prompt:Debug?|boolean;optional",
			wantPrompt:   "Debug?",
			wantType:     TypeBoolean,
			wantOptional: true,
		},
		{
			name:       "secret modifier",
			input:      "#prompt:Password?|string;secret",
			wantPrompt: "Password?",
			wantType:   TypeString,
			wantSecret: true,
		},
		{
			name:         "multiple modifiers",
			input:        "#prompt:API Key?|string;secret;optional;minlen:32",
			wantPrompt:   "API Key?",
			wantType:     TypeString,
			wantOptional: true,
			wantSecret:   true,
			wantConstraint: map[string]string{
				"minlen": "32",
			},
		},
		{
			name:       "object with format",
			input:      "#prompt:Config?|object;format:json",
			wantPrompt: "Config?",
			wantType:   TypeObject,
			wantConstraint: map[string]string{
				"format": "json",
			},
		},
		{
			name:       "numeric type",
			input:      "#prompt:Rate?|numeric;min:0;max:1",
			wantPrompt: "Rate?",
			wantType:   TypeNumeric,
			wantConstraint: map[string]string{
				"min": "0",
				"max": "1",
			},
		},
		{
			name:       "string with pattern",
			input:      "#prompt:Email?|string;pattern:^[a-z]+@[a-z]+\\.[a-z]+$",
			wantPrompt: "Email?",
			wantType:   TypeString,
			wantConstraint: map[string]string{
				"pattern": "^[a-z]+@[a-z]+\\.[a-z]+$",
			},
		},
		{
			name:    "invalid format",
			input:   "#prompt:Value?",
			wantErr: true,
		},
		{
			name:    "missing prompt prefix",
			input:   "prompt:Value?|string",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ann, err := ParseAnnotation(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, ann)

			assert.Equal(t, tt.wantPrompt, ann.PromptText)
			assert.Equal(t, tt.wantType, ann.Type)
			assert.Equal(t, tt.wantOptional, ann.IsOptional)
			assert.Equal(t, tt.wantSecret, ann.IsSecret)

			for name, value := range tt.wantConstraint {
				assert.Equal(t, value, ann.GetConstraint(name), "constraint %s", name)
			}
		})
	}
}

func TestParseEnvFile(t *testing.T) {
	input := `#krakenv:environments=local,testing,production
#krakenv:strict=false

# Database Configuration
DB_HOST=localhost #prompt:Database host?|string
DB_PORT=5432 #prompt:Database port?|int;min:1;max:65535
DB_PASSWORD= #prompt:Password?|string;secret

# Feature Flags
ENABLE_CACHE=true
`

	envFile, err := ParseEnvFileContent(input, "test.env.dist")
	require.NoError(t, err)
	require.NotNil(t, envFile)

	// Check config was parsed
	require.NotNil(t, envFile.Config)
	assert.Equal(t, []string{"local", "testing", "production"}, envFile.Config.Environments)
	assert.False(t, envFile.Config.Strict)

	// Check variables
	assert.Len(t, envFile.Variables, 4)

	// DB_HOST
	dbHost := envFile.GetVariable("DB_HOST")
	require.NotNil(t, dbHost)
	assert.Equal(t, "localhost", dbHost.Value)
	require.NotNil(t, dbHost.Annotation)
	assert.Equal(t, "Database host?", dbHost.Annotation.PromptText)
	assert.Equal(t, TypeString, dbHost.Annotation.Type)

	// DB_PORT
	dbPort := envFile.GetVariable("DB_PORT")
	require.NotNil(t, dbPort)
	assert.Equal(t, "5432", dbPort.Value)
	require.NotNil(t, dbPort.Annotation)
	assert.Equal(t, "1", dbPort.Annotation.GetConstraint("min"))
	assert.Equal(t, "65535", dbPort.Annotation.GetConstraint("max"))

	// DB_PASSWORD
	dbPassword := envFile.GetVariable("DB_PASSWORD")
	require.NotNil(t, dbPassword)
	assert.Empty(t, dbPassword.Value)
	require.NotNil(t, dbPassword.Annotation)
	assert.True(t, dbPassword.Annotation.IsSecret)

	// ENABLE_CACHE (no annotation)
	enableCache := envFile.GetVariable("ENABLE_CACHE")
	require.NotNil(t, enableCache)
	assert.Equal(t, "true", enableCache.Value)
	assert.Nil(t, enableCache.Annotation)
}

func TestParseEnvFile_EmptyFile(t *testing.T) {
	envFile, err := ParseEnvFileContent("", "empty.env")
	require.NoError(t, err)
	require.NotNil(t, envFile)
	assert.Empty(t, envFile.Variables)
}

func TestParseEnvFile_CommentsOnly(t *testing.T) {
	input := `# This is a comment
# Another comment
`
	envFile, err := ParseEnvFileContent(input, "comments.env")
	require.NoError(t, err)
	require.NotNil(t, envFile)
	assert.Empty(t, envFile.Variables)
	assert.Len(t, envFile.Comments, 2)
}

func TestParseEnvFile_LineNumbers(t *testing.T) {
	input := `# Comment on line 1
DB_HOST=localhost
# Comment on line 3
DB_PORT=5432
`
	envFile, err := ParseEnvFileContent(input, "test.env")
	require.NoError(t, err)

	dbHost := envFile.GetVariable("DB_HOST")
	require.NotNil(t, dbHost)
	assert.Equal(t, 2, dbHost.LineNumber)

	dbPort := envFile.GetVariable("DB_PORT")
	require.NotNil(t, dbPort)
	assert.Equal(t, 4, dbPort.LineNumber)
}

func TestParseEnvFile_PreservesOrder(t *testing.T) {
	input := `VAR_C=3
VAR_A=1
VAR_B=2
`
	envFile, err := ParseEnvFileContent(input, "test.env")
	require.NoError(t, err)

	assert.Len(t, envFile.Variables, 3)
	assert.Equal(t, "VAR_C", envFile.Variables[0].Name)
	assert.Equal(t, "VAR_A", envFile.Variables[1].Name)
	assert.Equal(t, "VAR_B", envFile.Variables[2].Name)
}

func TestAnnotation_GetConstraint(t *testing.T) {
	ann := &Annotation{
		Constraints: []Constraint{
			{Name: "min", Value: "1"},
			{Name: "max", Value: "100"},
		},
	}

	assert.Equal(t, "1", ann.GetConstraint("min"))
	assert.Equal(t, "100", ann.GetConstraint("max"))
	assert.Empty(t, ann.GetConstraint("pattern"))
}

func TestAnnotation_HasConstraint(t *testing.T) {
	ann := &Annotation{
		Constraints: []Constraint{
			{Name: "min", Value: "1"},
		},
	}

	assert.True(t, ann.HasConstraint("min"))
	assert.False(t, ann.HasConstraint("max"))
}

func TestParseAnnotation_UnknownConstraint(t *testing.T) {
	// Per FR-041: Unknown constraints should trigger warning and be ignored
	ann, err := ParseAnnotation("#prompt:Value?|string;unknown:value;minlen:1")
	require.NoError(t, err)
	require.NotNil(t, ann)

	// Known constraint should be present
	assert.Equal(t, "1", ann.GetConstraint("minlen"))

	// Unknown constraint should be ignored (not cause error)
	assert.Empty(t, ann.GetConstraint("unknown"))
}

func TestParseAnnotation_EmptyEnumOptions(t *testing.T) {
	// Per FR-042: Enum with empty options should be treated as string
	ann, err := ParseAnnotation("#prompt:Value?|enum;options:")
	require.NoError(t, err)
	require.NotNil(t, ann)

	// Type becomes string when enum has empty options
	assert.Equal(t, TypeString, ann.Type)
}

func TestParseEnvFile_WhitespaceValue(t *testing.T) {
	// Per FR-040: Whitespace-only values should be treated as empty
	input := `VAR_A=   
VAR_B=  hello  
`
	envFile, err := ParseEnvFileContent(input, "test.env")
	require.NoError(t, err)

	varA := envFile.GetVariable("VAR_A")
	require.NotNil(t, varA)
	assert.Empty(t, varA.Value) // Trimmed to empty

	varB := envFile.GetVariable("VAR_B")
	require.NotNil(t, varB)
	assert.Equal(t, "hello", varB.Value) // Trimmed
}

func TestParseEnvFile_DuplicateVariables(t *testing.T) {
	// Duplicate variables: last definition wins
	input := `DB_HOST=first
DB_HOST=second
`
	envFile, err := ParseEnvFileContent(input, "test.env")
	require.NoError(t, err)

	// Should only have one variable (last wins)
	dbHost := envFile.GetVariable("DB_HOST")
	require.NotNil(t, dbHost)
	assert.Equal(t, "second", dbHost.Value)
}

func TestAnnotation_EncodingConstraint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		encoding string
	}{
		{
			name:     "base64 encoding",
			input:    "#prompt:Certificate?|string;encoding:base64",
			encoding: "base64",
		},
		{
			name:     "heredoc encoding",
			input:    "#prompt:Script?|string;encoding:heredoc",
			encoding: "heredoc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ann, err := ParseAnnotation(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.encoding, ann.GetConstraint("encoding"))
		})
	}
}

func BenchmarkParseEnvFile(b *testing.B) {
	// Create a large file for benchmarking
	var builder strings.Builder
	builder.WriteString("#krakenv:environments=local,prod\n\n")
	for i := 0; i < 100; i++ {
		builder.WriteString("VAR_" + string(rune('A'+i%26)) + "_" + string(rune('0'+i/26)) + "=value" + string(rune('0'+i%10)) + " #prompt:Question?|string\n")
	}
	content := builder.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseEnvFileContent(content, "benchmark.env")
	}
}
