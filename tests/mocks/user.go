package mocks

import (
	"github.com/dhaifley/apid/auth"
	"github.com/dhaifley/apid/request"
	"github.com/dhaifley/apid/sqldb"
)

var TestUser = auth.User{
	UserID: request.FieldString{
		Set: true, Valid: true,
		Value: TestUUID,
	},
	Email: request.FieldString{
		Set: true, Valid: true,
		Value: "test@apid.io",
	},
	LastName: request.FieldString{
		Set: true, Valid: true,
		Value: "testLastName",
	},
	FirstName: request.FieldString{
		Set: true, Valid: true,
		Value: "testFirstName",
	},
	Status: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusActive,
	},
	Data: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"test": "test",
		},
	},
	CreatedByUser: &sqldb.UserData{},
	UpdatedByUser: &sqldb.UserData{},
}

type MockUserRow struct{}

func (m *MockUserRow) Scan(dest ...any) error {
	n := 0

	if v, ok := dest[n].(*int64); ok {
		*v = TestKey
		n++

		if len(dest) <= n {
			return nil
		}
	}

	if v, ok := dest[n].(*string); ok {
		*v = TestUser.UserID.Value
		n++

		if len(dest) <= n {
			return nil
		}
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.UserID
		n++

		if len(dest) <= n {
			return nil
		}
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.Email
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.LastName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.FirstName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.Status
		n++
	}

	if v, ok := dest[n].(*request.FieldJSON); ok {
		*v = TestUser.Data
		n++
	}

	if n >= len(dest) {
		return nil
	}

	if v, ok := dest[n].(*request.FieldTime); ok {
		*v = TestUser.CreatedAt
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.CreatedBy
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.CreatedByUser.UserID
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.CreatedByUser.Email
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.CreatedByUser.LastName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.CreatedByUser.FirstName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.CreatedByUser.Status
		n++
	}

	if v, ok := dest[n].(*request.FieldJSON); ok {
		*v = TestUser.CreatedByUser.Data
		n++
	}

	if v, ok := dest[n].(*request.FieldTime); ok {
		*v = TestUser.UpdatedAt
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.UpdatedBy
		n++
	}

	if n >= len(dest) {
		return nil
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.UpdatedByUser.UserID
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.UpdatedByUser.Email
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.UpdatedByUser.LastName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.UpdatedByUser.FirstName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.UpdatedByUser.Status
		n++
	}

	if v, ok := dest[n].(*request.FieldJSON); ok {
		*v = TestUser.UpdatedByUser.Data
	}

	return nil
}
