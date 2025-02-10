package mocks

import (
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/sqldb"
)

var TestUserData = sqldb.UserData{
	UserID: request.FieldString{
		Set: true, Valid: true,
		Value: TestID,
	},
	Email: request.FieldString{
		Set: true, Valid: true,
		Value: "test@test.com",
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
			"testData": "testData",
		},
	},
}

type MockUserDataRow struct{}

func (m *MockUserDataRow) Scan(dest ...any) error {
	n := 0

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUserData.UserID
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUserData.Email
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUserData.LastName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUserData.FirstName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestUserData.Status
		n++
	}

	if v, ok := dest[n].(*request.FieldJSON); ok {
		*v = TestUserData.Data
	}

	return nil
}
