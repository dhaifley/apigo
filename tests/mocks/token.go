package mocks

import (
	"time"

	"github.com/dhaifley/apigo/internal/auth"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/sqldb"
)

var TestToken = auth.Token{
	TokenID: request.FieldString{
		Set: true, Valid: true,
		Value: TestUUID,
	},
	Status: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusActive,
	},
	Expiration: request.FieldTime{
		Set: true, Valid: true,
		Value: time.Now().Add(time.Hour).Unix(),
	},
	CreatedByUser: &sqldb.UserData{},
	UpdatedByUser: &sqldb.UserData{},
}

type MockTokenRow struct{}

func (m *MockTokenRow) Scan(dest ...any) error {
	n := 0

	if len(dest) > 1 {
		if v, ok := dest[n].(*string); ok {
			*v = TestToken.Status.Value
			n++

			if len(dest) <= n {
				return nil
			}
		}
	}

	if v, ok := dest[n].(*string); ok {
		*v = TestToken.TokenID.Value
		n++

		if len(dest) <= n {
			return nil
		}
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestToken.TokenID
		n++

		if len(dest) <= n {
			return nil
		}
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestToken.Status
		n++
	}

	if v, ok := dest[n].(*request.FieldTime); ok {
		*v = TestToken.Expiration
		n++
	}

	if len(dest) == n {
		return nil
	}

	if v, ok := dest[n].(*request.FieldTime); ok {
		*v = TestToken.CreatedAt
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestToken.CreatedBy
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestToken.CreatedByUser.UserID
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestToken.CreatedByUser.Email
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestToken.CreatedByUser.LastName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestToken.CreatedByUser.FirstName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestToken.CreatedByUser.Status
		n++
	}

	if v, ok := dest[n].(*request.FieldJSON); ok {
		*v = TestToken.CreatedByUser.Data
		n++
	}

	if v, ok := dest[n].(*request.FieldTime); ok {
		*v = TestToken.UpdatedAt
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestToken.UpdatedBy
		n++
	}

	if n >= len(dest) {
		return nil
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestToken.UpdatedByUser.UserID
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestToken.UpdatedByUser.Email
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestToken.UpdatedByUser.LastName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestToken.UpdatedByUser.FirstName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestToken.UpdatedByUser.Status
		n++
	}

	if v, ok := dest[n].(*request.FieldJSON); ok {
		*v = TestToken.UpdatedByUser.Data
	}

	return nil
}

type MockTokenRows struct {
	row int
}

func (m *MockTokenRows) Err() error {
	return nil
}

func (m *MockTokenRows) Close() {
	return
}

func (m *MockTokenRows) Next() bool {
	m.row++

	return m.row <= 1
}

func (m *MockTokenRows) Scan(dest ...interface{}) error {
	r := &MockTokenRow{}

	return r.Scan(dest...)
}
