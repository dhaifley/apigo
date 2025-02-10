package resource_test

import (
	"testing"

	"github.com/dhaifley/apigo/internal/cache"
	"github.com/dhaifley/apigo/internal/resource"
	"github.com/dhaifley/apigo/tests/mocks"
)

func TestGetTags(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	svc := resource.NewService(nil, &mocks.MockResourceDB{}, nil, nil, nil, nil)

	res, err := svc.GetTags(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) <= 0 {
		t.Errorf("Expected tags length to be greater than 0")
	}

	if res["test"][0] != "test" {
		t.Errorf("Expected tag value: test, got: %v", res["test"][0])
	}
}

func TestGetResourceTags(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	svc := resource.NewService(nil, &mocks.MockResourceDB{}, nil, nil, nil, nil)

	res, err := svc.GetResourceTags(ctx, mocks.TestID)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) <= 0 {
		t.Errorf("Expected tags length to be greater than 0")
	}

	if res[0] != "test:test" {
		t.Errorf("Expected tag: test:test, got: %v", res[0])
	}
}

func TestAddResourceTags(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	svc := resource.NewService(nil, &mocks.MockResourceDB{}, nil, nil, nil, nil)

	res, err := svc.AddResourceTags(ctx, mocks.TestID, []string{"test:test"})
	if err != nil {
		t.Fatal(err)
	}

	if len(res) <= 0 {
		t.Errorf("Expected tags length to be greater than 0")
	}

	if res[0] != "test:test" {
		t.Errorf("Expected tag: test:test, got: %v", res[0])
	}
}

func TestDeleteResourceTags(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	svc := resource.NewService(nil, &mocks.MockResourceDB{}, nil, nil, nil, nil)

	err := svc.DeleteResourceTags(ctx, mocks.TestID, []string{"test:test"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateTagsMultiAssignment(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := resource.NewService(nil, &mocks.MockResourceDB{}, &mc,
		nil, nil, nil)

	res, err := svc.CreateTagsMultiAssignment(ctx,
		&mocks.TestTagsMultiAssignment)
	if err != nil {
		t.Fatal(err)
	}

	if res.Tags.Value[0] != mocks.TestTagsMultiAssignment.Tags.Value[0] {
		t.Errorf("Expected tag: %v, got: %v",
			mocks.TestTagsMultiAssignment.Tags.Value[0],
			res.Tags.Value[0])
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}
}

func TestDeleteTagsMultiAssignment(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := resource.NewService(nil, &mocks.MockResourceDB{}, &mc,
		nil, nil, nil)

	res, err := svc.DeleteTagsMultiAssignment(ctx,
		&mocks.TestTagsMultiAssignment)
	if err != nil {
		t.Fatal(err)
	}

	if res.Tags.Value[0] != mocks.TestTagsMultiAssignment.Tags.Value[0] {
		t.Errorf("Expected tag: %v, got: %v",
			mocks.TestTagsMultiAssignment.Tags.Value[0],
			res.Tags.Value[0])
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}
}
