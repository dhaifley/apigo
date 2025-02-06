package sqldb_test

import (
	"context"
	"testing"

	"github.com/dhaifley/apid/search"
	"github.com/dhaifley/apid/sqldb"
)

func TestNewQuery(t *testing.T) {
	t.Parallel()

	base := "SELECT * FROM user"

	req := &search.Query{
		Search: "and(id:1,email:test)",
	}

	primary := "email"

	fields := []*sqldb.Field{
		{
			Name:  "id",
			Type:  sqldb.FieldInt,
			Table: `"user"`,
		},
		{
			Name:    "email",
			Type:    sqldb.FieldString,
			Primary: true,
			Table:   `"user"`,
		},
		{
			Name:  "reminders",
			Type:  sqldb.FieldArray,
			Table: `"user"`,
		},
		{
			Name:  "tags",
			Type:  sqldb.FieldArray,
			Table: `"user"`,
		},
	}

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     &mockSQLConn{},
		Tx:     &mockSQLTrans{},
		Type:   sqldb.QuerySelect,
		Base:   base,
		Search: req,
		Fields: fields,
	})

	if q.Base != base {
		t.Errorf("Expecting base: %v, got: %v", base, q.Base)
	}

	if q.Field("email") != fields[1] {
		t.Errorf("Expecting fields: %v, got: %v", fields, q.Fields)
	}

	if q.Primary() != primary {
		t.Errorf("Expecting primary: %v, got: %v", primary, q.Primary())
	}

	if q.Type != sqldb.QuerySelect {
		t.Errorf("Expected type: SELECT, got: %v", q.Type)
	}
}

func TestQueryParse(t *testing.T) {
	base := "SELECT user.id FROM user"

	req := &search.Query{
		Search: `t"*"* and(id:1,test:test,also:*,notes:null,data.test:*,` +
			`data.array[1].test:test)`,
		Size:    10,
		From:    10,
		Order:   "-id",
		Summary: "id",
	}

	fields := []*sqldb.Field{
		{
			Name:  "id",
			Type:  sqldb.FieldInt,
			Table: `"user"`,
		},
		{
			Name:    "email",
			Type:    sqldb.FieldString,
			Primary: true,
			Table:   `"user"`,
		},
		{
			Name:  "notes",
			Type:  sqldb.FieldString,
			Table: `"user"`,
		},
		{
			Name:  "tags",
			Type:  sqldb.FieldArray,
			Table: `"user"`,
		},
		{
			Name:  "data",
			Type:  sqldb.FieldJSON,
			Table: `"user"`,
		},
	}

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     &mockSQLConn{},
		Tx:     &mockSQLTrans{},
		Type:   sqldb.QuerySelect,
		Base:   base,
		Search: req,
		Fields: fields,
	})

	if err := q.Parse(); err != nil {
		t.Fatal(err)
	}

	exp := "SELECT user.id FROM user WHERE " +
		"((\"user\".email LIKE $1) AND " +
		"((\"user\".id = $2) AND " +
		"($3 = ANY(\"user\".tags)) AND " +
		"(EXISTS (SELECT * FROM " +
		"(SELECT UNNEST(\"user\".tags)::text) val_tags(val) " +
		"WHERE val_tags.val LIKE $4 LIMIT 1)) AND " +
		"(\"user\".notes IS NULL) AND " +
		"(\"user\".data->>'test' LIKE $5) AND " +
		"(\"user\".data->'array'->1->>'test' = $6))) " +
		"GROUP BY \"user\".id " +
		"ORDER BY \"user\".id DESC LIMIT 11 OFFSET 10"

	if q.SQL != exp {
		t.Errorf("Expecting query: %v, got: %v", exp, q.SQL)
	}

	if q.Limit != 10 {
		t.Errorf("Expecting limit: 10, got: %v", q.Limit)
	}

	exp = "t*%"

	if q.Params[0] != exp {
		t.Errorf("Expecting param 0: %v, got: %v", exp, q.Params[0])
	}

	if q.Params[1] != int64(1) {
		t.Errorf("Expecting param 1: 1, got: %v", q.Params[1])
	}

	exp = "test:test"

	if q.Params[2] != exp {
		t.Errorf("Expecting param 2: %v, got: %v", exp, q.Params[2])
	}

	exp = "also:%"

	if q.Params[3] != exp {
		t.Errorf("Expecting param 3: %v, got: %v", exp, q.Params[3])
	}
}

func TestQueryNoParse(t *testing.T) {
	base := "SELECT account_url FROM accounts WHERE account_id = $1"

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     &mockSQLConn{},
		Tx:     &mockSQLTrans{},
		Type:   sqldb.QuerySelect,
		Base:   base,
		Params: []any{1},
	})

	q.Limit = 20

	if err := q.Parse(); err != nil {
		t.Fatal(err)
	}

	exp := "SELECT account_url FROM accounts WHERE account_id = $1 " +
		"LIMIT 21 OFFSET 0"

	if q.SQL != exp {
		t.Errorf("Expecting query: %v, got: %v", exp, q.SQL)
	}

	if q.Params[0] != 1 {
		t.Errorf("Expecting param 0: 1, got: %v", q.Params[0])
	}
}

func TestQueryExec(t *testing.T) {
	base := "SELECT * FROM user"

	fields := []*sqldb.Field{
		{
			Name: "user_id",
			Type: sqldb.FieldInt,
		},
	}

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     &mockSQLConn{},
		Tx:     &mockSQLTrans{},
		Type:   sqldb.QuerySelect,
		Base:   base,
		Fields: fields,
	})

	if err := q.Parse(); err != nil {
		t.Fatal(err)
	}

	if _, err := q.Exec(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestQueryParseDelete(t *testing.T) {
	base := "DELETE FROM user"

	fields := []*sqldb.Field{
		{
			Name:  "user_id",
			Type:  sqldb.FieldInt,
			Table: `"user"`,
		},
		{
			Name:    "email",
			Type:    sqldb.FieldString,
			Primary: true,
			Table:   `"user"`,
		},
	}

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     &mockSQLConn{},
		Tx:     &mockSQLTrans{},
		Type:   sqldb.QueryDelete,
		Base:   base,
		Search: &search.Query{},
		Fields: fields,
	})

	if err := q.Parse(); err == nil {
		t.Fatal("Expecting error for delete without query parameters.")
	}
}

func TestQueryQuery(t *testing.T) {
	base := "SELECT * FROM user"

	fields := []*sqldb.Field{
		{
			Name: "user_id",
			Type: sqldb.FieldInt,
		},
	}

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     &mockSQLConn{},
		Tx:     &mockSQLTrans{},
		Type:   sqldb.QuerySelect,
		Base:   base,
		Fields: fields,
	})

	if err := q.Parse(); err != nil {
		t.Fatal(err)
	}

	if _, err := q.Query(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestQueryQueryRow(t *testing.T) {
	base := "SELECT * FROM user"

	fields := []*sqldb.Field{
		{
			Name: "user_id",
			Type: sqldb.FieldInt,
		},
	}

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     &mockSQLConn{},
		Tx:     &mockSQLTrans{},
		Type:   sqldb.QuerySelect,
		Base:   base,
		Fields: fields,
	})

	if err := q.Parse(); err != nil {
		t.Fatal(err)
	}

	if _, err := q.QueryRow(context.Background()); err != nil {
		t.Error(err)
	}
}

func TestQueryInsert(t *testing.T) {
	base := "INSERT INTO user () VALUES () " +
		"ON CONFLICT DO UPDATE SET RETURNING id"

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:   &mockSQLConn{},
		Type: sqldb.QueryInsert,
		Base: base,
		Sets: []string{
			"test", "another", "account_id", "created_by", "updated_by",
		},
		Fields: []*sqldb.Field{{
			Name: "another",
			Type: sqldb.FieldTime,
		}, {
			Name:  "account_id",
			Table: "account",
			Key:   "account_key",
			Type:  sqldb.FieldString,
		}, {
			Name:  "created_by",
			Table: "create_user",
			Type:  sqldb.FieldString,
		}, {
			Name:  "updated_by",
			Table: "update_user",
			Type:  sqldb.FieldString,
		}},
		Params: []any{1, "test", testUUID, testUUID, testUUID},
	})

	if err := q.Parse(); err != nil {
		t.Fatal(err)
	}

	exp := `INSERT INTO user ` +
		`(test, another, account_key, created_by, updated_by) VALUES ` +
		`($1, to_timestamp($2), ` +
		`(SELECT account_key FROM account WHERE account_id = $3), ` +
		`(SELECT user_key FROM "user" WHERE user_id = $4), ` +
		`(SELECT user_key FROM "user" WHERE user_id = $5)) ` +
		`ON CONFLICT DO UPDATE SET test = $1, another = to_timestamp($2), ` +
		`account_key = (SELECT account_key FROM account ` +
		`WHERE account_id = $3), ` +
		`created_by = (SELECT user_key FROM "user" WHERE user_id = $4), ` +
		`updated_by = (SELECT user_key FROM "user" WHERE user_id = $5) ` +
		`RETURNING id`

	if q.SQL != exp {
		t.Errorf("Expected query: %v, got: %v", exp, q.SQL)
	}

	if _, err := q.QueryRow(context.Background()); err != nil {
		t.Error(err)
	}
}

func TestQueryUpdate(t *testing.T) {
	base := "UPDATE user SET WHERE id = $1 RETURNING id"

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:   &mockSQLConn{},
		Type: sqldb.QueryUpdate,
		Base: base,
		Sets: []string{
			"test", "another", "account_id", "created_by", "updated_by",
		},
		Fields: []*sqldb.Field{{
			Name: "another",
			Type: sqldb.FieldTime,
		}, {
			Name:  "account_id",
			Table: "account",
			Key:   "account_key",
			Type:  sqldb.FieldString,
		}, {
			Name:  "created_by",
			Table: "create_user",
			Type:  sqldb.FieldString,
		}, {
			Name:  "updated_by",
			Table: "update_user",
			Type:  sqldb.FieldString,
		}},
		Params: []any{1, "test", testUUID, testUUID, testUUID},
	})

	if err := q.Parse(); err != nil {
		t.Fatal(err)
	}

	exp := `UPDATE user SET test = $1, another = to_timestamp($2), ` +
		`account_key = (SELECT account_key FROM account ` +
		`WHERE account_id = $3), ` +
		`created_by = (SELECT user_key FROM "user" WHERE user_id = $4), ` +
		`updated_by = (SELECT user_key FROM "user" WHERE user_id = $5) ` +
		`WHERE id = $1 RETURNING id`

	if q.SQL != exp {
		t.Errorf("Expected query: %v, got: %v", exp, q.SQL)
	}

	if _, err := q.QueryRow(context.Background()); err != nil {
		t.Error(err)
	}
}
