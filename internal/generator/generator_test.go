package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theburrowhub/krakenv/internal/parser"
)

func TestNewGenerator(t *testing.T) {
	distFile := &parser.EnvFile{Path: ".env.dist"}
	gen := NewGenerator(distFile, ".env.local")

	assert.Equal(t, distFile, gen.DistFile)
	assert.Equal(t, ".env.local", gen.TargetPath)
	assert.Nil(t, gen.TargetFile)
}

func TestGenerator_GetVariablesToPrompt(t *testing.T) {
	distFile := &parser.EnvFile{
		Variables: []parser.Variable{
			{
				Name:       "WITH_ANNOTATION",
				Value:      "",
				Annotation: &parser.Annotation{PromptText: "Question?", Type: parser.TypeString},
			},
			{
				Name:       "WITH_DEFAULT",
				Value:      "default",
				Annotation: &parser.Annotation{PromptText: "Has default?", Type: parser.TypeString},
			},
			{
				Name:  "NO_ANNOTATION",
				Value: "value",
			},
		},
	}

	gen := NewGenerator(distFile, ".env.local")

	toPrompt := gen.GetVariablesToPrompt()

	// WITH_DEFAULT has a value so shouldn't be prompted
	// WITH_ANNOTATION has no value so should be prompted
	// NO_ANNOTATION has no annotation so shouldn't be prompted
	assert.Len(t, toPrompt, 1)
	assert.Equal(t, "WITH_ANNOTATION", toPrompt[0].Name)
}

func TestGenerator_GetVariablesToPrompt_WithExistingTarget(t *testing.T) {
	distFile := &parser.EnvFile{
		Variables: []parser.Variable{
			{
				Name:       "VAR_A",
				Value:      "",
				Annotation: &parser.Annotation{PromptText: "A?", Type: parser.TypeString},
			},
			{
				Name:       "VAR_B",
				Value:      "",
				Annotation: &parser.Annotation{PromptText: "B?", Type: parser.TypeString},
			},
		},
	}

	targetFile := &parser.EnvFile{
		Variables: []parser.Variable{
			{Name: "VAR_A", Value: "existing_value"},
		},
	}

	gen := NewGenerator(distFile, ".env.local")
	gen.TargetFile = targetFile

	toPrompt := gen.GetVariablesToPrompt()

	// VAR_A has existing value, VAR_B doesn't
	assert.Len(t, toPrompt, 1)
	assert.Equal(t, "VAR_B", toPrompt[0].Name)
}

func TestGenerator_MergeVariables(t *testing.T) {
	distFile := &parser.EnvFile{
		Variables: []parser.Variable{
			{Name: "VAR_A", Value: "dist_default"},
			{Name: "VAR_B", Value: ""},
			{Name: "VAR_C", Value: ""},
		},
	}

	targetFile := &parser.EnvFile{
		Variables: []parser.Variable{
			{Name: "VAR_B", Value: "target_value"},
		},
	}

	gen := NewGenerator(distFile, ".env.local")
	gen.TargetFile = targetFile

	userValues := map[string]string{
		"VAR_C": "user_value",
	}

	result := gen.MergeVariables(userValues)

	assert.Len(t, result, 3)
	assert.Equal(t, "dist_default", result[0].Value) // VAR_A: from dist
	assert.Equal(t, "target_value", result[1].Value) // VAR_B: from target
	assert.Equal(t, "user_value", result[2].Value)   // VAR_C: from user
}

func TestGenerator_WriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, ".env.local")

	distFile := &parser.EnvFile{
		Variables: []parser.Variable{
			{Name: "DB_HOST", Value: "localhost"},
			{Name: "DB_PORT", Value: "5432"},
		},
	}

	gen := NewGenerator(distFile, targetPath)

	err := gen.WriteFile(distFile.Variables)
	require.NoError(t, err)

	// Read back and verify
	content, err := os.ReadFile(targetPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "DB_HOST=localhost")
	assert.Contains(t, string(content), "DB_PORT=5432")
}

func TestGenerator_WriteFile_WithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, ".env.local")

	distFile := &parser.EnvFile{
		Config: &parser.KrakenvConfig{
			Environments: []string{"local", "prod"},
			Strict:       true,
		},
		Variables: []parser.Variable{
			{Name: "VAR", Value: "value"},
		},
	}

	gen := NewGenerator(distFile, targetPath)

	err := gen.WriteFile(distFile.Variables)
	require.NoError(t, err)

	content, err := os.ReadFile(targetPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "#krakenv:environments=local,prod")
	assert.Contains(t, string(content), "#krakenv:strict=true")
}

func TestGenerator_WriteFile_KeepAnnotations(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, ".env.local")

	distFile := &parser.EnvFile{
		Variables: []parser.Variable{
			{
				Name:       "PORT",
				Value:      "8080",
				Annotation: &parser.Annotation{PromptText: "Port?", Type: parser.TypeInt},
			},
		},
	}

	gen := NewGenerator(distFile, targetPath)
	gen.KeepAnnotations = true

	err := gen.WriteFile(distFile.Variables)
	require.NoError(t, err)

	content, err := os.ReadFile(targetPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "#prompt:Port?|int")
}

func TestGenerator_WriteFile_StripAnnotations(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, ".env.local")

	distFile := &parser.EnvFile{
		Variables: []parser.Variable{
			{
				Name:       "PORT",
				Value:      "8080",
				Annotation: &parser.Annotation{PromptText: "Port?", Type: parser.TypeInt},
			},
		},
	}

	gen := NewGenerator(distFile, targetPath)
	gen.KeepAnnotations = false // Default

	err := gen.WriteFile(distFile.Variables)
	require.NoError(t, err)

	content, err := os.ReadFile(targetPath)
	require.NoError(t, err)

	assert.NotContains(t, string(content), "#prompt:")
	assert.Contains(t, string(content), "PORT=8080")
}

func TestGenerate_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create distributable
	distContent := `#krakenv:environments=local
DB_HOST=localhost #prompt:Host?|string
DB_PORT= #prompt:Port?|int
`
	distPath := filepath.Join(tmpDir, ".env.dist")
	err := os.WriteFile(distPath, []byte(distContent), 0644)
	require.NoError(t, err)

	targetPath := filepath.Join(tmpDir, ".env.local")

	values := map[string]string{
		"DB_PORT": "5432",
	}

	result, err := Generate(distPath, targetPath, values, false)
	require.NoError(t, err)

	assert.True(t, result.Created)
	assert.Equal(t, targetPath, result.Path)
	assert.Equal(t, 2, result.Variables)

	// Verify output
	content, err := os.ReadFile(targetPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "DB_HOST=localhost")
	assert.Contains(t, string(content), "DB_PORT=5432")
}

func TestGenerate_UpdateExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create distributable
	distContent := `DB_HOST= #prompt:Host?|string
DB_PORT= #prompt:Port?|int
`
	distPath := filepath.Join(tmpDir, ".env.dist")
	err := os.WriteFile(distPath, []byte(distContent), 0644)
	require.NoError(t, err)

	// Create existing target
	existingContent := `DB_HOST=existing_host
`
	targetPath := filepath.Join(tmpDir, ".env.local")
	err = os.WriteFile(targetPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	values := map[string]string{
		"DB_PORT": "3306",
	}

	result, err := Generate(distPath, targetPath, values, false)
	require.NoError(t, err)

	assert.False(t, result.Created) // Updated, not created

	content, err := os.ReadFile(targetPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "DB_HOST=existing_host") // Preserved
	assert.Contains(t, string(content), "DB_PORT=3306")          // New value
}
