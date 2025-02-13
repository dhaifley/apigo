package repo_test

import (
	"testing"

	"github.com/dhaifley/apigo/internal/repo"
)

func TestGitHubClient(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name:    "Valid GitHub URL",
			repoURL: "github://user:token@owner/repo/path#ref",
		},
		{
			name:    "Missing token",
			repoURL: "github://user@owner/repo/path#ref",
			wantErr: true,
		},
		{
			name:    "Invalid URL format",
			repoURL: "github://owner/repo",
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
