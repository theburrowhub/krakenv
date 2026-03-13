package secrets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFetcher implements Fetcher for testing.
type mockFetcher struct {
	secrets map[string]string
	err     error
	calls   []string
}

func (m *mockFetcher) FetchSecret(_ context.Context, resourceName string) (string, error) {
	m.calls = append(m.calls, resourceName)
	if m.err != nil {
		return "", m.err
	}
	if v, ok := m.secrets[resourceName]; ok {
		return v, nil
	}
	return "", nil
}

func (m *mockFetcher) Close() error { return nil }

var _ Fetcher = (*mockFetcher)(nil)

func TestBuildResourceName(t *testing.T) {
	tests := []struct {
		name       string
		project    string
		secretName string
		version    string
		wantName   string
		wantErr    bool
	}{
		{
			name:       "all fields explicit",
			project:    "my-project",
			secretName: "my-secret",
			version:    "3",
			wantName:   "projects/my-project/secrets/my-secret/versions/3",
		},
		{
			name:       "v-prefixed version normalised",
			project:    "my-project",
			secretName: "my-secret",
			version:    "v7",
			wantName:   "projects/my-project/secrets/my-secret/versions/7",
		},
		{
			name:       "empty version defaults to latest",
			project:    "my-project",
			secretName: "my-secret",
			version:    "",
			wantName:   "projects/my-project/secrets/my-secret/versions/latest",
		},
		{
			name:       "different projects for different secrets",
			project:    "payments-project",
			secretName: "stripe-key",
			version:    "v2",
			wantName:   "projects/payments-project/secrets/stripe-key/versions/2",
		},
		{
			name:       "missing project returns error",
			project:    "",
			secretName: "my-secret",
			version:    "1",
			wantErr:    true,
		},
		{
			name:       "missing secret name returns error",
			project:    "my-project",
			secretName: "",
			version:    "1",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildResourceName(tt.project, tt.secretName, tt.version)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, got)
		})
	}
}
