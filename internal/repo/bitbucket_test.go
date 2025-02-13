package repo_test

import (
	"testing"

	"github.com/dhaifley/apigo/internal/repo"
)

func TestBitBucketClient(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name:    "Valid BitBucket URL",
			repoURL: "bitbucket://user:token@owner/repo/path#ref",
		},
		{
			name:    "Missing credentials",
			repoURL: "bitbucket://owner/repo/path#ref",
			wantErr: true,
		},
		{
			name:    "Invalid URL format",
			repoURL: "bitbucket://owner",
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
