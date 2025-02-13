package sqldb_test

import (
	"net/url"
	"strings"
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

func TestSummaryData(t *testing.T) {
	t.Parallel()

	fields := []*sqldb.Field{
		{
			Name: "string_field",
			Type: sqldb.FieldString,
		},
		{
			Name: "int_field",
			Type: sqldb.FieldInt,
		},
		{
			Name: "float_field",
			Type: sqldb.FieldFloat,
		},
		{
			Name: "bool_field",
			Type: sqldb.FieldBool,
		},
		{
			Name: "time_field",
			Type: sqldb.FieldTime,
		},
		{
			Name: "array_field",
			Type: sqldb.FieldArray,
		},
		{
			Name: "json_field",
			Type: sqldb.FieldJSON,
		},
	}

	sd := make(sqldb.SummaryData)

	query := &search.Query{
		Summary: "string_field,int_field,float_field,bool_field," +
			"time_field,array_field,json_field",
	}

	dest := sd.ScanDest(fields, query)

	if len(dest) != 8 { // 7 fields + count
		t.Errorf("Expected 8 scan destinations, got %d", len(dest))
	}

	// Verify each destination type
	if _, ok := dest[0].(*string); !ok {
		t.Error("Expected string destination for string_field")
	}

	if _, ok := dest[1].(*int64); !ok {
		t.Error("Expected int64 destination for int_field")
	}

	if _, ok := dest[2].(*float64); !ok {
		t.Error("Expected float64 destination for float_field")
	}

	if _, ok := dest[3].(*bool); !ok {
		t.Error("Expected bool destination for bool_field")
	}

	if _, ok := dest[4].(*int64); !ok {
		t.Error("Expected int64 destination for time_field")
	}

	if _, ok := dest[5].(*[]any); !ok {
		t.Error("Expected []any destination for array_field")
	}

	if _, ok := dest[6].(*map[string]any); !ok {
		t.Error("Expected map[string]any destination for json_field")
	}

	if _, ok := dest[7].(*int64); !ok {
		t.Error("Expected int64 destination for count")
	}
}

func TestFieldString(t *testing.T) {
	t.Parallel()

	f := &sqldb.Field{
		Name: "test",
		Type: sqldb.FieldString,
	}

	str := f.String()
	if !strings.Contains(str, `"name":"test"`) {
		t.Errorf("Expected field string to contain name, got %s", str)
	}
}
