package resource_test

import (
	"testing"

	"github.com/dhaifley/apigo/internal/cache"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/resource"
	"github.com/dhaifley/apigo/internal/sqldb"
	"github.com/pashagolub/pgxmock/v4"
)

const TestTag = "test:test"

func mockTagRows(mock pgxmock.PgxCommonIface) *pgxmock.Rows {
	return mock.NewRows([]string{"tag"}).
		AddRow(TestTag)
}

func TestGetTags(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, nil, nil, nil, nil)

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM tag").
		WillReturnRows(mockTagRows(mock))

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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestGetResourceTags(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, nil, nil, nil, nil)

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM tag").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockTagRows(mock))

	res, err := svc.GetResourceTags(ctx, TestID)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) <= 0 {
		t.Errorf("Expected tags length to be greater than 0")
	}

	if res[0] != "test:test" {
		t.Errorf("Expected tag: test:test, got: %v", res[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestAddResourceTags(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, nil, nil, nil, nil)

	mockTransaction(mock)

	args := make([]any, 4)

	for i := 0; i < 4; i++ {
		args[i] = pgxmock.AnyArg()
	}

	mock.ExpectQuery("INSERT INTO tag_obj").
		WithArgs(args...).WillReturnRows(mockTagRows(mock))

	res, err := svc.AddResourceTags(ctx, TestID, []string{"test:test"})
	if err != nil {
		t.Fatal(err)
	}

	if len(res) <= 0 {
		t.Errorf("Expected tags length to be greater than 0")
	}

	if res[0] != "test:test" {
		t.Errorf("Expected tag: test:test, got: %v", res[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestDeleteResourceTags(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, nil, nil, nil, nil)

	mockTransaction(mock)

	args := make([]any, 3)

	for i := 0; i < 3; i++ {
		args[i] = pgxmock.AnyArg()
	}

	mock.ExpectQuery("DELETE FROM tag_obj").
		WithArgs(args...).WillReturnRows(mockTagRows(mock))

	mockTransaction(mock)

	args = args[:2]

	mock.ExpectQuery("DELETE FROM tag").
		WithArgs(args...).WillReturnRows(mockTagRows(mock))

	err = svc.DeleteResourceTags(ctx, TestID, []string{"test:test"})
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

var TestTagsMultiAssignment = resource.TagsMultiAssignment{
	Tags: request.FieldStringArray{
		Set: true, Valid: true,
		Value: []string{"test:test"},
	},
	ResourceSelector: request.FieldString{
		Set: true, Valid: true,
		Value: "and(name:*)",
	},
}

func TestCreateTagsMultiAssignment(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM resource").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockResourceKeyRows(mock))

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM resource").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockResourceRows(mock))

	mockTransaction(mock)

	args := make([]any, 4)

	for i := 0; i < 4; i++ {
		args[i] = pgxmock.AnyArg()
	}

	mock.ExpectQuery("INSERT INTO tag_obj").
		WithArgs(args...).WillReturnRows(mockTagRows(mock))

	res, err := svc.CreateTagsMultiAssignment(ctx,
		&TestTagsMultiAssignment)
	if err != nil {
		t.Fatal(err)
	}

	if res.Tags.Value[0] != TestTagsMultiAssignment.Tags.Value[0] {
		t.Errorf("Expected tag: %v, got: %v",
			TestTagsMultiAssignment.Tags.Value[0],
			res.Tags.Value[0])
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestDeleteTagsMultiAssignment(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM resource").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockResourceKeyRows(mock))

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM resource").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockResourceRows(mock))

	mockTransaction(mock)

	args := make([]any, 3)

	for i := 0; i < 3; i++ {
		args[i] = pgxmock.AnyArg()
	}

	mock.ExpectQuery("DELETE FROM tag_obj").
		WithArgs(args...).WillReturnRows(mockTagRows(mock))

	mockTransaction(mock)

	args = args[:2]

	mock.ExpectQuery("DELETE FROM tag").
		WithArgs(args...).WillReturnRows(mockTagRows(mock))

	res, err := svc.DeleteTagsMultiAssignment(ctx,
		&TestTagsMultiAssignment)
	if err != nil {
		t.Fatal(err)
	}

	if res.Tags.Value[0] != TestTagsMultiAssignment.Tags.Value[0] {
		t.Errorf("Expected tag: %v, got: %v",
			TestTagsMultiAssignment.Tags.Value[0],
			res.Tags.Value[0])
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}
