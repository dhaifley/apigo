package repo_test

import (
	"context"
	"testing"

	"github.com/dhaifley/apid/repo"
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
