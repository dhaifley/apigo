package mocks

import (
	"github.com/dhaifley/apid/request"
	"github.com/dhaifley/apid/resource"
)

type MockTagRow struct{}

func (m *MockTagRow) Scan(dest ...any) error {
	n := 0

	if len(dest) == n {
		return nil
	}

	if v, ok := dest[n].(*string); ok {
		*v = "test:test"
	}

	return nil
}

type MockTagRows struct {
	row int
}

func (m *MockTagRows) Err() error {
	return nil
}

func (m *MockTagRows) Close() {
	return
}

func (m *MockTagRows) Next() bool {
	m.row++

	return m.row <= 1
}

func (m *MockTagRows) Scan(dest ...interface{}) error {
	r := &MockTagRow{}

	return r.Scan(dest...)
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
