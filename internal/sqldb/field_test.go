package sqldb_test

import (
	"net/url"
	"testing"

	"github.com/dhaifley/apigo/internal/search"
	"github.com/dhaifley/apigo/internal/sqldb"
)

func TestParseFieldOptions(t *testing.T) {
	t.Parallel()

	options, err := sqldb.ParseFieldOptions(url.Values{
		"user_details": []string{"true"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !options.Contains(sqldb.OptUserDetails) {
		t.Errorf("Expected: %v, got: %v", sqldb.OptUserDetails, options)
	}
}

func TestSelectFields(t *testing.T) {
	t.Parallel()

	exp := "test"

	v := sqldb.SelectFields(exp, []*sqldb.Field{{
		Name:  exp + "_key",
		Table: exp,
		Type:  sqldb.FieldInt,
	}}, nil,
		[]sqldb.FieldOption{sqldb.OptUserDetails})

	exp = `SELECT
	test.test_key AS test_test_key
FROM test
`

	if v != exp {
		t.Errorf("Expected: %v, got: %v", exp, v)
	}

	v = sqldb.SelectFields("test", []*sqldb.Field{{
		Name:  "status",
		Table: "test",
		Type:  sqldb.FieldString,
	}}, &search.Query{
		Summary: "status",
	}, nil)

	exp = `SELECT
	test.status AS test_status,
	COUNT(*) AS count
FROM test
`

	if v != exp {
		t.Errorf("Expected: %v, got: %v", exp, v)
	}
}

func TestSearchFields(t *testing.T) {
	t.Parallel()

	exp := "test"

	v := sqldb.SearchFields(exp, []*sqldb.Field{{
		Name:  exp + "_key",
		Table: exp,
		Type:  sqldb.FieldInt,
	}})

	exp = `SELECT
	test.test_key AS test_test_key
FROM test
`

	if v != exp {
		t.Errorf("Expected: %v, got: %v", exp, v)
	}
}

func TestReturningFields(t *testing.T) {
	t.Parallel()

	exp := "test"

	v := sqldb.ReturningFields(exp, []*sqldb.Field{{
		Name:  exp + "_key",
		Table: exp,
		Type:  sqldb.FieldInt,
	}}, []sqldb.FieldOption{sqldb.OptUserDetails})

	exp = "\n" + `RETURNING
	test.test_key AS test_test_key` + "\n"

	if v != exp {
		t.Errorf("Expected: %v, got: %v", exp, v)
	}
}
