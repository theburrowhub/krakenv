package envfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.env")
	content := `DB_HOST=localhost #prompt:Host?|string
DB_PORT=5432 #prompt:Port?|int;min:1;max:65535
`
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	// Parse
	env, err := Parse(path)
	require.NoError(t, err)
	require.NotNil(t, env)

	assert.Len(t, env.Variables, 2)
	assert.Equal(t, "DB_HOST", env.Variables[0].Name)
	assert.Equal(t, "localhost", env.Variables[0].Value)
}

func TestParseContent(t *testing.T) {
	content := `DB_HOST=localhost
DB_PORT=5432
`
	env, err := ParseContent(content, "inline.env")
	require.NoError(t, err)
	require.NotNil(t, env)

	assert.Len(t, env.Variables, 2)
}

func TestValidate(t *testing.T) {
	ann := &Annotation{
		Type: TypeInt,
		Constraints: []Constraint{
			{Name: "min", Value: "1"},
			{Name: "max", Value: "100"},
		},
	}

	// Valid
	err := Validate("50", ann)
	assert.NoError(t, err)

	// Invalid
	err = Validate("200", ann)
	assert.Error(t, err)

	// Not an int
	err = Validate("abc", ann)
	assert.Error(t, err)
}

func TestValidateFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create dist
	distContent := `PORT= #prompt:Port?|int;min:1;max:65535
HOST=localhost #prompt:Host?|string
`
	distPath := filepath.Join(tmpDir, ".env.dist")
	err := os.WriteFile(distPath, []byte(distContent), 0644)
	require.NoError(t, err)

	// Create target
	targetContent := `PORT=8080
HOST=example.com
`
	targetPath := filepath.Join(tmpDir, ".env.local")
	err = os.WriteFile(targetPath, []byte(targetContent), 0644)
	require.NoError(t, err)

	// Validate
	result, err := ValidateFile(distPath, targetPath)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestFormatAnnotation(t *testing.T) {
	ann := &Annotation{
		PromptText:  "Enter port?",
		Type:        TypeInt,
		Constraints: []Constraint{{Name: "min", Value: "1"}},
		IsOptional:  true,
	}

	formatted := FormatAnnotation(ann)
	assert.Contains(t, formatted, "#prompt:Enter port?|int")
	assert.Contains(t, formatted, "min:1")
	assert.Contains(t, formatted, "optional")
}

func TestFormatVariable(t *testing.T) {
	v := Variable{
		Name:  "PORT",
		Value: "8080",
		Annotation: &Annotation{
			PromptText: "Port?",
			Type:       TypeInt,
		},
	}

	// Without annotation
	line := FormatVariable(v, false)
	assert.Equal(t, "PORT=8080", line)

	// With annotation
	line = FormatVariable(v, true)
	assert.Contains(t, line, "PORT=8080")
	assert.Contains(t, line, "#prompt:Port?|int")
}
