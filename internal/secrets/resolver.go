package secrets

import (
	"context"
	"fmt"

	"github.com/theburrowhub/krakenv/internal/parser"
)

// ResolveFromEnvFile inspects all variables in envFile and fetches values for
// those annotated with the gcp-secret modifier.
//
// Each such variable must declare all three GCP fields in its annotation:
//
//	gcp-secret-project:PROJECT
//	gcp-secret-name:SECRET_NAME
//	gcp-secret-version:VERSION   (optional — defaults to "latest")
//
// Returns a map of variable name → secret value for every variable resolved.
// Returns an error if any annotated variable is missing required fields or if
// the GCP API call fails.
func ResolveFromEnvFile(ctx context.Context, fetcher Fetcher, envFile *parser.EnvFile) (map[string]string, error) {
	resolved := make(map[string]string)

	for _, v := range envFile.Variables {
		if v.Annotation == nil || !v.Annotation.IsGCPSecret {
			continue
		}

		project := v.Annotation.GetConstraint("gcp-secret-project")
		name := v.Annotation.GetConstraint("gcp-secret-name")
		version := v.Annotation.GetConstraint("gcp-secret-version")

		resourceName, err := BuildResourceName(project, name, version)
		if err != nil {
			return nil, fmt.Errorf("variable %s: %w", v.Name, err)
		}

		value, err := fetcher.FetchSecret(ctx, resourceName)
		if err != nil {
			return nil, fmt.Errorf("variable %s: %w", v.Name, err)
		}

		resolved[v.Name] = value
	}

	return resolved, nil
}

// HasGCPSecrets reports whether any variable in envFile has the gcp-secret modifier.
func HasGCPSecrets(envFile *parser.EnvFile) bool {
	for _, v := range envFile.Variables {
		if v.Annotation != nil && v.Annotation.IsGCPSecret {
			return true
		}
	}
	return false
}
