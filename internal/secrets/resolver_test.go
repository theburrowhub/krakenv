package secrets

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theburrowhub/krakenv/internal/parser"
)

func buildEnvFile(vars []parser.Variable) *parser.EnvFile {
	return &parser.EnvFile{
		Path:      ".env.dist",
		Variables: vars,
	}
}

func gcpAnnotation(project, name, version string) *parser.Annotation {
	constraints := []parser.Constraint{
		{Name: "gcp-secret-project", Value: project},
		{Name: "gcp-secret-name", Value: name},
	}
	if version != "" {
		constraints = append(constraints, parser.Constraint{Name: "gcp-secret-version", Value: version})
	}
	return &parser.Annotation{
		PromptText:  "Enter value?",
		IsGCPSecret: true,
		Constraints: constraints,
	}
}

func TestHasGCPSecrets(t *testing.T) {
	t.Run("no variables", func(t *testing.T) {
		assert.False(t, HasGCPSecrets(buildEnvFile(nil)))
	})

	t.Run("variable without annotation", func(t *testing.T) {
		ef := buildEnvFile([]parser.Variable{{Name: "FOO", Value: "bar"}})
		assert.False(t, HasGCPSecrets(ef))
	})

	t.Run("annotated variable without gcp-secret modifier", func(t *testing.T) {
		ef := buildEnvFile([]parser.Variable{
			{Name: "DB_PASS", Annotation: &parser.Annotation{PromptText: "Password?", IsSecret: true}},
		})
		assert.False(t, HasGCPSecrets(ef))
	})

	t.Run("variable with gcp-secret modifier", func(t *testing.T) {
		ef := buildEnvFile([]parser.Variable{
			{Name: "API_KEY", Annotation: gcpAnnotation("my-project", "api-key", "")},
		})
		assert.True(t, HasGCPSecrets(ef))
	})
}

func TestResolveFromEnvFile(t *testing.T) {
	ctx := context.Background()

	t.Run("resolves multiple gcp-secret variables from different projects", func(t *testing.T) {
		apiKeyResource := "projects/payments-project/secrets/stripe-key/versions/2"
		dbPassResource := "projects/infra-project/secrets/db-password/versions/latest"

		mock := &mockFetcher{
			secrets: map[string]string{
				apiKeyResource: "sk_live_abc123",
				dbPassResource: "supersecure",
			},
		}
		ef := buildEnvFile([]parser.Variable{
			{Name: "API_KEY", Annotation: gcpAnnotation("payments-project", "stripe-key", "v2")},
			{Name: "DB_PASSWORD", Annotation: gcpAnnotation("infra-project", "db-password", "")},
			{Name: "PLAIN_VAR", Value: "plain"},
		})

		resolved, err := ResolveFromEnvFile(ctx, mock, ef)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{
			"API_KEY":     "sk_live_abc123",
			"DB_PASSWORD": "supersecure",
		}, resolved)
		assert.Equal(t, []string{apiKeyResource, dbPassResource}, mock.calls)
	})

	t.Run("returns empty map when no gcp-secret modifiers", func(t *testing.T) {
		mock := &mockFetcher{}
		ef := buildEnvFile([]parser.Variable{{Name: "FOO", Value: "bar"}})

		resolved, err := ResolveFromEnvFile(ctx, mock, ef)
		require.NoError(t, err)
		assert.Empty(t, resolved)
		assert.Empty(t, mock.calls)
	})

	t.Run("error when gcp-secret-project missing", func(t *testing.T) {
		ann := &parser.Annotation{
			PromptText:  "Value?",
			IsGCPSecret: true,
			Constraints: []parser.Constraint{
				{Name: "gcp-secret-name", Value: "my-secret"},
			},
		}
		ef := buildEnvFile([]parser.Variable{{Name: "MY_VAR", Annotation: ann}})

		_, err := ResolveFromEnvFile(ctx, &mockFetcher{}, ef)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "MY_VAR")
		assert.Contains(t, err.Error(), "gcp-secret-project")
	})

	t.Run("error when gcp-secret-name missing", func(t *testing.T) {
		ann := &parser.Annotation{
			PromptText:  "Value?",
			IsGCPSecret: true,
			Constraints: []parser.Constraint{
				{Name: "gcp-secret-project", Value: "my-project"},
			},
		}
		ef := buildEnvFile([]parser.Variable{{Name: "MY_VAR", Annotation: ann}})

		_, err := ResolveFromEnvFile(ctx, &mockFetcher{}, ef)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "MY_VAR")
		assert.Contains(t, err.Error(), "gcp-secret-name")
	})

	t.Run("propagates fetcher error with variable name context", func(t *testing.T) {
		fetchErr := errors.New("permission denied")
		mock := &mockFetcher{err: fetchErr}
		ef := buildEnvFile([]parser.Variable{
			{Name: "MY_VAR", Annotation: gcpAnnotation("my-project", "my-secret", "1")},
		})

		_, err := ResolveFromEnvFile(ctx, mock, ef)
		require.Error(t, err)
		assert.ErrorIs(t, err, fetchErr)
		assert.Contains(t, err.Error(), "MY_VAR")
	})

	t.Run("skips variables without gcp-secret modifier", func(t *testing.T) {
		resource := "projects/my-project/secrets/real-secret/versions/latest"
		mock := &mockFetcher{secrets: map[string]string{resource: "value"}}
		ef := buildEnvFile([]parser.Variable{
			{
				Name: "PROMPT_ONLY",
				Annotation: &parser.Annotation{
					PromptText:  "Enter value?",
					Constraints: []parser.Constraint{{Name: "minlen", Value: "1"}},
				},
			},
			{Name: "GCP_VAR", Annotation: gcpAnnotation("my-project", "real-secret", "")},
		})

		resolved, err := ResolveFromEnvFile(ctx, mock, ef)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"GCP_VAR": "value"}, resolved)
		assert.Equal(t, []string{resource}, mock.calls)
	})
}
