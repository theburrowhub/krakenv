package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theburrowhub/krakenv/internal/parser"
)

func TestValidateInt(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		min     string
		max     string
		wantErr bool
	}{
		{"valid int", "42", "", "", false},
		{"valid zero", "0", "", "", false},
		{"valid negative", "-10", "", "", false},
		{"valid with min", "5", "1", "", false},
		{"valid with max", "5", "", "10", false},
		{"valid in range", "5", "1", "10", false},
		{"at min boundary", "1", "1", "10", false},
		{"at max boundary", "10", "1", "10", false},
		{"below min", "0", "1", "10", true},
		{"above max", "11", "1", "10", true},
		{"not a number", "abc", "", "", true},
		{"float not allowed", "3.14", "", "", true},
		{"empty string", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ann := &parser.Annotation{Type: parser.TypeInt}
			if tt.min != "" {
				ann.Constraints = append(ann.Constraints, parser.Constraint{Name: "min", Value: tt.min})
			}
			if tt.max != "" {
				ann.Constraints = append(ann.Constraints, parser.Constraint{Name: "max", Value: tt.max})
			}

			err := ValidateValue(tt.value, ann)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNumeric(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		min     string
		max     string
		wantErr bool
	}{
		{"valid float", "3.14", "", "", false},
		{"valid int as numeric", "42", "", "", false},
		{"valid negative", "-3.14", "", "", false},
		{"valid zero", "0", "", "", false},
		{"valid with min", "0.5", "0", "", false},
		{"valid with max", "0.5", "", "1", false},
		{"valid in range", "0.5", "0", "1", false},
		{"below min", "-0.1", "0", "1", true},
		{"above max", "1.1", "0", "1", true},
		{"not a number", "abc", "", "", true},
		{"empty string", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ann := &parser.Annotation{Type: parser.TypeNumeric}
			if tt.min != "" {
				ann.Constraints = append(ann.Constraints, parser.Constraint{Name: "min", Value: tt.min})
			}
			if tt.max != "" {
				ann.Constraints = append(ann.Constraints, parser.Constraint{Name: "max", Value: tt.max})
			}

			err := ValidateValue(tt.value, ann)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateString(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		minlen  string
		maxlen  string
		pattern string
		wantErr bool
	}{
		{"valid string", "hello", "", "", "", false},
		{"empty allowed by default", "", "", "", "", false},
		{"valid with minlen", "hello", "1", "", "", false},
		{"valid with maxlen", "hi", "", "10", "", false},
		{"valid in length range", "hello", "1", "10", "", false},
		{"below minlen", "", "1", "", "", true},
		{"above maxlen", "hello world", "", "5", "", true},
		{"valid pattern", "abc123", "", "", "^[a-z0-9]+$", false},
		{"invalid pattern", "ABC123", "", "", "^[a-z0-9]+$", true},
		{"email pattern", "test@example.com", "", "", "^[^@]+@[^@]+\\.[^@]+$", false},
		{"invalid email", "not-an-email", "", "", "^[^@]+@[^@]+\\.[^@]+$", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ann := &parser.Annotation{Type: parser.TypeString}
			if tt.minlen != "" {
				ann.Constraints = append(ann.Constraints, parser.Constraint{Name: "minlen", Value: tt.minlen})
			}
			if tt.maxlen != "" {
				ann.Constraints = append(ann.Constraints, parser.Constraint{Name: "maxlen", Value: tt.maxlen})
			}
			if tt.pattern != "" {
				ann.Constraints = append(ann.Constraints, parser.Constraint{Name: "pattern", Value: tt.pattern})
			}

			err := ValidateValue(tt.value, ann)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEnum(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		options string
		wantErr bool
	}{
		{"valid first option", "dev", "dev,staging,prod", false},
		{"valid middle option", "staging", "dev,staging,prod", false},
		{"valid last option", "prod", "dev,staging,prod", false},
		{"invalid option", "test", "dev,staging,prod", true},
		{"case sensitive", "Dev", "dev,staging,prod", true},
		{"empty value", "", "dev,staging,prod", true},
		{"single option valid", "only", "only", false},
		{"whitespace trimmed", "dev", "dev , staging , prod", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ann := &parser.Annotation{
				Type: parser.TypeEnum,
				Constraints: []parser.Constraint{
					{Name: "options", Value: tt.options},
				},
			}

			err := ValidateValue(tt.value, ann)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateBoolean(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"true lowercase", "true", false},
		{"false lowercase", "false", false},
		{"TRUE uppercase", "TRUE", false},
		{"FALSE uppercase", "FALSE", false},
		{"True mixed", "True", false},
		{"yes", "yes", false},
		{"no", "no", false},
		{"YES uppercase", "YES", false},
		{"1", "1", false},
		{"0", "0", false},
		{"on", "on", false},
		{"off", "off", false},
		{"invalid", "maybe", true},
		{"empty", "", true},
		{"number other than 0/1", "2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ann := &parser.Annotation{Type: parser.TypeBoolean}

			err := ValidateValue(tt.value, ann)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateObject(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		format  string
		wantErr bool
	}{
		{"valid json object", `{"key":"value"}`, "json", false},
		{"valid json array", `[1, 2, 3]`, "json", false},
		{"valid json string", `"hello"`, "json", false},
		{"invalid json", `{key: value}`, "json", true},
		{"valid yaml", "key: value", "yaml", false},
		{"valid yaml list", "- item1\n- item2", "yaml", false},
		{"empty json", "", "json", true},
		{"empty yaml", "", "yaml", true},
		{"default format is json", `{"key":"value"}`, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ann := &parser.Annotation{Type: parser.TypeObject}
			if tt.format != "" {
				ann.Constraints = append(ann.Constraints, parser.Constraint{Name: "format", Value: tt.format})
			}

			err := ValidateValue(tt.value, ann)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOptional(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		optional bool
		wantErr  bool
	}{
		{"required with value", "hello", false, false},
		{"required empty", "", false, true},
		{"optional with value", "hello", true, false},
		{"optional empty", "", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ann := &parser.Annotation{
				Type:       parser.TypeString,
				IsOptional: tt.optional,
				Constraints: []parser.Constraint{
					{Name: "minlen", Value: "1"},
				},
			}

			err := ValidateValue(tt.value, ann)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateVariable(t *testing.T) {
	v := &parser.Variable{
		Name:       "TEST_VAR",
		Value:      "invalid",
		LineNumber: 5,
		Annotation: &parser.Annotation{
			Type: parser.TypeInt,
		},
	}

	err := ValidateVariable(v)
	require.NotNil(t, err)
	assert.Equal(t, "TEST_VAR", err.Variable)
	assert.Equal(t, 5, err.LineNumber)
	assert.Equal(t, ErrorInvalidType, err.Type)
}

func TestValidateVariable_NoAnnotation(t *testing.T) {
	v := &parser.Variable{
		Name:  "TEST_VAR",
		Value: "anything",
	}

	err := ValidateVariable(v)
	assert.Nil(t, err) // No annotation = no validation
}

func TestValidateEnvFile(t *testing.T) {
	envFile := &parser.EnvFile{
		Path: "test.env",
		Variables: []parser.Variable{
			{
				Name:       "VALID_INT",
				Value:      "42",
				LineNumber: 1,
				Annotation: &parser.Annotation{Type: parser.TypeInt},
			},
			{
				Name:       "INVALID_INT",
				Value:      "abc",
				LineNumber: 2,
				Annotation: &parser.Annotation{Type: parser.TypeInt},
			},
			{
				Name:       "VALID_STRING",
				Value:      "hello",
				LineNumber: 3,
				Annotation: &parser.Annotation{Type: parser.TypeString},
			},
			{
				Name:  "NO_ANNOTATION",
				Value: "anything",
			},
		},
	}

	result := ValidateEnvFile(envFile)
	assert.False(t, result.Valid)
	assert.Equal(t, 1, result.ErrorCount())
	assert.Equal(t, "INVALID_INT", result.Errors[0].Variable)
}

func TestValidateEnvFile_AllValid(t *testing.T) {
	envFile := &parser.EnvFile{
		Path: "test.env",
		Variables: []parser.Variable{
			{
				Name:       "PORT",
				Value:      "8080",
				Annotation: &parser.Annotation{Type: parser.TypeInt},
			},
			{
				Name:       "HOST",
				Value:      "localhost",
				Annotation: &parser.Annotation{Type: parser.TypeString},
			},
		},
	}

	result := ValidateEnvFile(envFile)
	assert.True(t, result.Valid)
	assert.Equal(t, 0, result.ErrorCount())
}

func TestValidationResult_FormatErrors(t *testing.T) {
	result := NewValidationResult()
	result.AddError(ValidationError{
		Variable:   "PORT",
		LineNumber: 5,
		Message:    "Expected integer, got \"abc\"",
		Suggestion: "Enter a valid port number",
		Example:    "8080",
		Type:       ErrorInvalidType,
	})

	output := result.FormatErrors("test.env")
	assert.Contains(t, output, "VALIDATION FAILED")
	assert.Contains(t, output, "Line 5: PORT")
	assert.Contains(t, output, "Expected integer")
	assert.Contains(t, output, "Enter a valid port")
	assert.Contains(t, output, "8080")
}
