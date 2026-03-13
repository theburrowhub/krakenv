// Package secrets provides clients for fetching secret values from remote stores.
package secrets

import (
	"context"
	"fmt"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// Fetcher is the interface for fetching secret values from a remote store.
// Implementations must be safe for concurrent use.
type Fetcher interface {
	// FetchSecret retrieves a secret value by its canonical GCP resource name:
	// "projects/PROJECT/secrets/SECRET_NAME/versions/VERSION".
	FetchSecret(ctx context.Context, resourceName string) (string, error)
	// Close releases any resources held by the fetcher.
	Close() error
}

// GCPClient fetches secrets from Google Cloud Secret Manager using
// Application Default Credentials (ADC).
type GCPClient struct {
	client *secretmanager.Client
}

// NewGCPClient creates a GCP Secret Manager client using ADC.
func NewGCPClient(ctx context.Context) (*GCPClient, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Secret Manager client: %w", err)
	}
	return &GCPClient{client: client}, nil
}

// FetchSecret retrieves the payload of a secret version.
// resourceName must be the canonical GCP path:
// "projects/PROJECT/secrets/SECRET_NAME/versions/VERSION".
func (c *GCPClient) FetchSecret(ctx context.Context, resourceName string) (string, error) {
	result, err := c.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: resourceName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to access secret %q: %w", resourceName, err)
	}
	return string(result.Payload.Data), nil
}

// Close releases the underlying gRPC connection.
func (c *GCPClient) Close() error {
	return c.client.Close()
}

// BuildResourceName constructs the canonical GCP Secret Manager resource name
// from the three explicit fields declared in the .env.dist annotation.
//
// project  — value of gcp-secret-project constraint (required)
// name     — value of gcp-secret-name constraint (required)
// version  — value of gcp-secret-version constraint (optional; defaults to "latest")
//
// The version accepts both numeric ("3") and v-prefixed ("v3") notation.
func BuildResourceName(project, name, version string) (string, error) {
	if project == "" {
		return "", fmt.Errorf("gcp-secret-project is required")
	}
	if name == "" {
		return "", fmt.Errorf("gcp-secret-name is required")
	}

	v := strings.TrimPrefix(version, "v") // normalize "v3" → "3"
	if v == "" {
		v = "latest"
	}

	return fmt.Sprintf("projects/%s/secrets/%s/versions/%s", project, name, v), nil
}
