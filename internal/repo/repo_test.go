package repo_test

import (
	"context"
	"testing"

	"github.com/dhaifley/apigo/internal/repo"
)

func mockContext() context.Context {
	return context.WithValue(context.Background(), 5,
		"11223344-5566-7788-9900-aabbccddeeff")
}

func TestRepo(t *testing.T) {
	ctx := mockContext()

	cli, err := repo.NewClient("test://user:token@owner/repo/path#ref",
		nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := cli.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	if _, err = cli.List(ctx, "/"); err != nil {
		t.Fatal(err)
	}

	res, err := cli.ListAll(ctx, "/")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := cli.Get(ctx, res[0].Path); err != nil {
		t.Fatal(err)
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name:    "Valid test URL",
			repoURL: "test://user:token@owner/repo/path#ref",
		},
		{
			name:    "Empty URL",
			repoURL: "",
			wantErr: true,
		},
		{
			name:    "Invalid URL format",
			repoURL: "not@a-url",
			wantErr: true,
		},
		{
			name:    "Unsupported scheme",
			repoURL: "unsupported://test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, err := repo.NewClient(tt.repoURL, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if err != nil {
				return
			}

			if cli == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestItem(t *testing.T) {
	item := repo.Item{
		Path:       "/test/path",
		Attributes: []string{"attr1", "attr2"},
		Mimetype:   "text/plain",
		Size:       100,
		Type:       "file",
		Commit:     "abc123",
	}

	if item.Path != "/test/path" {
		t.Errorf("Item.Path = %v, want %v", item.Path, "/test/path")
	}

	if len(item.Attributes) != 2 {
		t.Errorf("len(Item.Attributes) = %v, want %v", len(item.Attributes), 2)
	}

	if item.Mimetype != "text/plain" {
		t.Errorf("Item.Mimetype = %v, want %v", item.Mimetype, "text/plain")
	}

	if item.Size != 100 {
		t.Errorf("Item.Size = %v, want %v", item.Size, 100)
	}

	if item.Type != "file" {
		t.Errorf("Item.Type = %v, want %v", item.Type, "file")
	}

	if item.Commit != "abc123" {
		t.Errorf("Item.Commit = %v, want %v", item.Commit, "abc123")
	}
}
