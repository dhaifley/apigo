package repo_test

import (
	"testing"

	"github.com/dhaifley/apigo/internal/repo"
)

func TestGitClient(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name:    "Valid Git URL",
			repoURL: "https://user:token@github.com/repo.git",
		},
		{
			name:    "Invalid Git URL",
			repoURL: "git@foo:invalid-url",
			wantErr: true,
		},
		{
			name:    "Invalid scheme",
			repoURL: "foo://example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.NewClient(tt.repoURL, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
		})
	}
}
