package mocks

import (
	"github.com/dhaifley/apid/internal/auth"
	"github.com/dhaifley/apid/internal/request"
)

var TestAccount = auth.Account{
	AccountID: request.FieldString{
		Set: true, Valid: true,
		Value: TestID,
	},
	Name: request.FieldString{
		Set: true, Valid: true,
		Value: "testAccount",
	},
	Status: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusActive,
	},
	StatusData: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"last_error": "test",
		},
	},
	Repo: request.FieldString{
		Set: true, Valid: true,
		Value: "test",
	},
	RepoStatus: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusActive,
	},
	Secret: request.FieldString{
		Set: true, Valid: true,
		Value: "test",
	},
	Data: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"test": "test",
		},
	},
}

type MockAccountRow struct{}

func (m *MockAccountRow) Scan(dest ...any) error {
	n := 0

	if len(dest) == 3 {
		if v, ok := dest[n].(*request.FieldString); ok {
			*v = TestAccount.Repo
			n++
		}

		if v, ok := dest[n].(*request.FieldString); ok {
			*v = TestAccount.RepoStatus
			n++
		}

		if v, ok := dest[n].(*request.FieldJSON); ok {
			*v = TestAccount.RepoStatusData
		}

		return nil
	}

	if v, ok := dest[n].(*int64); ok {
		*v = TestKey
		n++

		if len(dest) <= n {
			return nil
		}
	}

	if v, ok := dest[n].(*string); ok {
		*v = TestAccount.AccountID.Value
		n++

		if len(dest) <= n {
			return nil
		}
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestAccount.AccountID
		n++

		if len(dest) <= n {
			return nil
		}
	}

	if v, ok := dest[n].(**string); ok {
		*v = new(string)
		**v = TestUUID
		n++

		if len(dest) <= n {
			return nil
		}

		return nil
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestAccount.Name
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestAccount.Status
		n++
	}

	if v, ok := dest[n].(*request.FieldJSON); ok {
		*v = TestAccount.StatusData
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestAccount.Repo
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestAccount.RepoStatus
		n++
	}

	if v, ok := dest[n].(*request.FieldJSON); ok {
		*v = TestAccount.RepoStatusData
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestAccount.Secret
		n++
	}

	if v, ok := dest[n].(*request.FieldJSON); ok {
		*v = TestAccount.Data
		n++
	}

	if v, ok := dest[n].(*request.FieldTime); ok {
		*v = TestAccount.CreatedAt
		n++
	}

	if v, ok := dest[n].(*request.FieldTime); ok {
		*v = TestAccount.UpdatedAt
	}

	return nil
}

type MockAccountRows struct {
	row int
}

func (m *MockAccountRows) Err() error {
	return nil
}

func (m *MockAccountRows) Close() {
	return
}

func (m *MockAccountRows) Next() bool {
	m.row++

	return m.row <= 1
}

func (m *MockAccountRows) Scan(dest ...interface{}) error {
	r := &MockAccountRow{}

	return r.Scan(dest...)
}
