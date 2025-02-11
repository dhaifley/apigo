package mocks

import (
	"github.com/dhaifley/apigo/internal/auth"
	"github.com/dhaifley/apigo/internal/request"
)

var TestUser = auth.User{
	UserID: request.FieldString{
		Set: true, Valid: true,
		Value: TestUUID,
	},
	Email: request.FieldString{
		Set: true, Valid: true,
		Value: "test@apigo.io",
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
	Scopes: request.FieldString{
		Set: true, Valid: true,
		Value: request.ScopeSuperUser,
	},
	Data: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"test": "test",
		},
	},
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

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.Scopes
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

	if v, ok := dest[n].(*request.FieldTime); ok {
		*v = TestUser.UpdatedAt
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUser.UpdatedBy
	}

	return nil
}
