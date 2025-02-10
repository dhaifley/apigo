package sqldb

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dhaifley/apigo/internal/config"
	"github.com/dhaifley/apigo/internal/errors"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/search"
)

// Query values are used to build SQL queries for search operations.
type Query struct {
	Config   *config.Config `json:"-"`
	DB       SQLDB          `json:"db"`
	Tx       SQLTX          `json:"tx,omitempty"`
	Type     QueryType      `json:"type"`
	SQL      string         `json:"sql"`
	Base     string         `json:"base"`
	Search   *search.Query  `json:"search,omitempty"`
	Fields   []*Field       `json:"search_fields,omitempty"`
	Sets     []string       `json:"set_fields,omitempty"`
	Params   []any          `json:"params,omitempty"`
	Limit    int64          `json:"limit"`
	count    int64          `json:"-"`
	setStart int64          `json:"-"`
}

// QueryType is an enum type describing the type of SQL query.
type QueryType string

// Supported query type values.
const (
	QuerySelect = QueryType("SELECT")
	QueryInsert = QueryType("INSERT")
	QueryUpdate = QueryType("UPDATE")
	QueryDelete = QueryType("DELETE")
	QueryExec   = QueryType("EXEC")
)

// QueryOptions values contain options when creating a new query.
type QueryOptions struct {
	Config *config.Config `json:"-"`
	DB     SQLDB          `json:"db"`
	Tx     SQLTX          `json:"tx,omitempty"`
	Type   QueryType      `json:"type"`
	Base   string         `json:"base"`
	Search *search.Query  `json:"search,omitempty"`
	Fields []*Field       `json:"fields,omitempty"`
	Sets   []string       `json:"set,omitempty"`
	Params []any          `json:"params,omitempty"`
}

// NewQuery creates an initializes a new query value.
func NewQuery(opts *QueryOptions) *Query {
	if opts == nil {
		return nil
	}

	cfg := opts.Config
	if cfg == nil {
		cfg = config.NewDefault()
	}

	return &Query{
		Config:   cfg,
		DB:       opts.DB,
		Tx:       opts.Tx,
		Type:     opts.Type,
		Base:     opts.Base,
		Search:   opts.Search,
		Fields:   opts.Fields,
		Sets:     opts.Sets,
		Params:   opts.Params,
		SQL:      "",
		Limit:    0,
		count:    int64(len(opts.Params)),
		setStart: int64(len(opts.Params)-len(opts.Sets)) + 1,
	}
}

// Field retrieves a query search field value by name.
func (q *Query) Field(name string) *Field {
	for _, f := range q.Fields {
		if name == f.Name {
			return f
		}

		for _, a := range f.Search {
			if name == a {
				return f
			}
		}
	}

	return nil
}

// Primary retrieves the name of the primary search field.
func (q *Query) Primary() string {
	for _, f := range q.Fields {
		if f.Primary {
			return f.Name
		}
	}

	return ""
}

// escapeWildcards converts and escapes wildcard characters.
func (q *Query) escapeWildcards(s string) string {
	str := strings.ReplaceAll(s, "%", `\%`)
	str = strings.ReplaceAll(str, "_", `\_`)
	str = strings.ReplaceAll(str, "\\?", "«")
	str = strings.ReplaceAll(str, "\\*", "»")
	str = strings.ReplaceAll(str, "*", "%")
	str = strings.ReplaceAll(str, "?", "_")
	str = strings.ReplaceAll(str, "«", "?")
	str = strings.ReplaceAll(str, "»", "*")
	str = strings.ReplaceAll(str, "÷", "?")
	str = strings.ReplaceAll(str, "°", "*")

	return str
}

// containsWildcards determines whether a string contains wildcard characters.
func (q *Query) containsWildcards(s string) bool {
	return strings.ContainsAny(s, "*?÷°")
}

// addParam validates and appends a search field parameter value.
func (q *Query) addParam(f *Field, value string) error {
	value = strings.TrimSpace(value)

	if strings.ToLower(value) == "null" {
		return nil
	}

	var v any

	var err error

	switch f.Type {
	case FieldString:
		switch {
		case q.containsWildcards(value):
			v = q.escapeWildcards(value)
		case value == "":
			v = "%"
		default:
			v = value
		}
	case FieldInt:
		switch {
		case q.containsWildcards(value):
			v = q.escapeWildcards(value)
		case value == "":
			v = "%"
		default:
			v, err = strconv.ParseInt(value, 10, 64)
			if err != nil {
				return errors.Wrap(err, errors.ErrInvalidRequest,
					"unable to parse integer parameter",
					"param", value)
			}
		}
	case FieldFloat:
		switch {
		case q.containsWildcards(value):
			v = q.escapeWildcards(value)
		case value == "":
			v = "%"
		default:
			v, err = strconv.ParseFloat(value, 64)
			if err != nil {
				return errors.Wrap(err, errors.ErrInvalidRequest,
					"unable to parse float param",
					"param", value)
			}
		}
	case FieldBool:
		switch {
		case q.containsWildcards(value):
			return nil
		case value == "":
			v = "true"
		default:
			v, err = strconv.ParseBool(value)
			if err != nil {
				return errors.Wrap(err, errors.ErrInvalidRequest,
					"unable to parse bool param",
					"param", value)
			}
		}
	case FieldTime:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to parse time param",
				"param", value)
		}

		if f.Op == "<" {
			i++
		} else if f.Op == ">" {
			i--
		}

		v = i
	case FieldArray:
		switch {
		case q.containsWildcards(value):
			v = q.escapeWildcards(value)
		case value == "":
			v = "%"
		default:
			v = value
		}
	case FieldJSON:
		switch {
		case q.containsWildcards(value):
			v = q.escapeWildcards(value)
		case value == "":
			v = "%"
		default:
			v = value
		}
	default:
		return errors.New(errors.ErrInvalidRequest,
			"invalid search field type",
			"field", f)
	}

	if v != nil {
		q.Params = append(q.Params, v)
		q.count++
	}

	return nil
}

// formatParam returns a where cause expression for a search field value.
func (q *Query) formatParam(f *Field,
	jsonExpr string,
	op FieldOperator,
	value string,
) (string, error) {
	if strings.ToLower(value) == "null" {
		return fmt.Sprintf("(%s.%s IS NULL)", f.Table, f.Name), nil
	}

	name := f.Name

	expr := f.Expr

	param := fmt.Sprintf("$%d", q.count)

	switch f.Type {
	case FieldString:
		if op == "" {
			if q.containsWildcards(value) {
				op = OpLike
			} else {
				op = OpEq
			}
		}
	case FieldBool:
		if q.containsWildcards(value) {
			if expr != "" {
				return fmt.Sprintf("(%s = true OR %s = false)",
					expr, expr), nil
			}

			if f.Table == "" {
				return fmt.Sprintf("(%s = true OR %s = false)",
					name, name), nil
			}

			return fmt.Sprintf("(%s.%s = true OR %s.%s = false)",
				f.Table, name, f.Table, name), nil
		}

		op = OpEq
	case FieldTime:
		param = "to_timestamp(" + param + ")"
	case FieldInt, FieldFloat:
		if op == OpLike {
			name += "::text"

			if expr != "" {
				expr += "::text"
			}
		}

		if op == "" {
			op = OpEq
		}
	case FieldArray:
		res := ""

		if q.containsWildcards(value) {
			stub := "(EXISTS (SELECT * FROM "

			if f.Tags {
				res += stub + `(SELECT
				tag_obj.tag_key || ':' || tag_obj.tag_val AS tag
			FROM tag_obj
			WHERE tag_obj.status = '` + request.StatusActive + `'
				AND tag_obj.tag_type = '` + strings.Trim(f.Table, `"`) + `'
				AND tag_obj.tag_obj_id = ` + f.Table + `.` +
					strings.Trim(f.Table, `"`) + `_id::TEXT)
				AND (tag_obj.tag_key || ':' || tag_obj.tag_val) LIKE $` +
					strconv.FormatInt(q.count, 10) + ` LIMIT 1) tags LIMIT 1))`
			} else if f.Expr != "" {
				res += fmt.Sprintf(stub+
					"(SELECT UNNEST(%s)::text) expr_%d(val) "+
					"WHERE expr_%d.val LIKE $%d LIMIT 1))",
					f.Expr, q.count,
					q.count, q.count)
			} else if f.Table == "" {
				res += fmt.Sprintf(stub+
					"(SELECT UNNEST(%s)::text) val_%s(val) "+
					"WHERE val_%s.val LIKE $%d LIMIT 1))",
					f.Name, f.Name,
					f.Name, q.count)
			} else {
				res += fmt.Sprintf(stub+
					"(SELECT UNNEST(%s.%s)::text) val_%s(val) "+
					"WHERE val_%s.val LIKE $%d LIMIT 1))",
					f.Table, f.Name, f.Name,
					f.Name, q.count)
			}
		} else {
			if f.Tags {
				exp := `(SELECT
				ARRAY_AGG(tag_obj.tag_key || ':' || tag_obj.tag_val) AS tags
			FROM tag_obj
			WHERE tag_obj.status = '` + request.StatusActive + `'
				AND tag_obj.tag_type = '` + strings.Trim(f.Table, `"`) + `'
				AND tag_obj.tag_obj_id = ` + f.Table + `.` +
					strings.Trim(f.Table, `"`) + `_id::TEXT)::TEXT[]`

				res += fmt.Sprintf("(%s = ANY(%s))", param, exp)
			} else if f.Expr != "" {
				res += fmt.Sprintf("(%s = ANY(%s))", param, f.Expr)
			} else if f.Table == "" {
				res += fmt.Sprintf("(%s = ANY(%s))", param, f.Name)
			} else {
				res += fmt.Sprintf("(%s = ANY(%s.%s))",
					param, f.Table, f.Name)
			}
		}

		return res, nil
	case FieldJSON:
		if op == "" {
			if q.containsWildcards(value) {
				op = OpLike
			} else {
				op = OpEq
			}
		}

		if jsonExpr != "" {
			jop := "->"

			if !strings.Contains(jsonExpr, "->") {
				jop += ">"
			}

			if expr != "" {
				expr += jop + jsonExpr
			}

			name += jop + jsonExpr
		}
	default:
		return "", errors.New(errors.ErrInvalidRequest,
			"invalid search field type",
			"field", f)
	}

	if f.Op != "" {
		op = f.Op
	}

	if expr != "" {
		return fmt.Sprintf("(%s %s %s)", expr, op, param), nil
	}

	if f.Table == "" {
		return fmt.Sprintf("(%s %s %s)", name, op, param), nil
	}

	return fmt.Sprintf("(%s.%s %s %s)", f.Table, name, op, param), nil
}

// parseSearchNode returns a SQL where clause expression for a single search
// syntax tree node.
func (q *Query) parseSearchNode(node *search.QueryNode,
) (string, error) {
	if node == nil {
		return "", nil
	}

	switch node.Op {
	case search.OpMatch:
		val := node.Val

		op := OpEq

		if node.ValRE != "" {
			val = node.ValRE
			op = OpRE
		} else if q.containsWildcards(val) || val == "" {
			op = OpLike
		}

		var field *Field

		jsonExpr := ""

		if strings.Contains(node.Cat, ".") {
			parts := strings.Split(node.Cat, ".")

			if len(parts) > 1 {
				jsonExpr = "'" + strings.ReplaceAll(
					strings.Join(parts[1:], "."), ".", "'->'")

				jsonExpr = strings.ReplaceAll(strings.ReplaceAll(
					strings.ReplaceAll(jsonExpr,
						"[", "'->"), "]'->'", "->'"), "]", "") + "'"

				if i := strings.LastIndex(jsonExpr, "->"); i >= 0 {
					jsonExpr = jsonExpr[:i] + "->>" + jsonExpr[i+2:]
				}

				field = q.Field(parts[0])
			} else {
				field = q.Field(node.Cat)
			}
		} else {
			field = q.Field(node.Cat)
		}

		if field == nil {
			// Attempt to use the term as a tag search.
			if field = q.Field("tags"); field == nil {
				return "", errors.New(errors.ErrInvalidRequest,
					"invalid search term",
					"term", node.Cat)
			}

			op = OpAny

			if val != "" {
				val = node.Cat + ":" + val
			} else {
				val = node.Cat
			}
		}

		if err := q.addParam(field, val); err != nil {
			return "", err
		}

		return q.formatParam(field, jsonExpr, op, val)
	case search.OpAnd, search.OpOr, search.OpNot:
		nodes := []string{}

		for _, n := range node.Nodes {
			if str, err := q.parseSearchNode(n); err != nil {
				return "", err
			} else if str != "" {
				nodes = append(nodes, str)
			}
		}

		if len(nodes) > 0 {
			if node.Op == search.OpNot {
				return "(NOT " + nodes[0] + ")", nil
			}

			return "(" + strings.Join(nodes, " "+
				strings.ToUpper(node.Op.String())+" ") + ")", nil
		}
	}

	return "", nil
}

// parseSearch parses the search query string value.
func (q *Query) parseSearch() error {
	var (
		err error
		ast *search.QueryTree
	)

	qp := search.NewParser(bytes.NewBufferString(q.Search.Search))

	qp.Primary = q.Primary()

	if ast, err = qp.Parse(); err != nil {
		return errors.Wrap(err, errors.ErrInvalidRequest,
			"invalid search query",
			"search", q.Search.Search)
	}

	if sql, err := q.parseSearchNode(ast.Root); err != nil {
		return err
	} else if sql != "" {
		if !strings.Contains(q.SQL, "WHERE") {
			sql = "WHERE " + sql
		} else {
			sql = "AND " + sql
		}

		switch {
		case strings.Contains(q.SQL, "RETURNING"):
			q.SQL = strings.Replace(q.SQL, "RETURNING", sql+" RETURNING", 1)
		case strings.Contains(q.SQL, "GROUP BY"):
			q.SQL = strings.Replace(q.SQL, "GROUP BY", sql+" GROUP BY", 1)
		case strings.Contains(q.SQL, "UNION"):
			q.SQL = strings.Replace(q.SQL, "UNION", sql+" UNION", 1)
		case strings.Contains(q.SQL, "UNION"):
			q.SQL = strings.Replace(q.SQL, "UNION", sql+" UNION", 1)
		case strings.Contains(q.SQL, "ORDER BY"):
			q.SQL = strings.Replace(q.SQL, "ORDER BY", sql+" ORDER BY", 1)
		case strings.Contains(q.SQL, "LIMIT"):
			q.SQL = strings.Replace(q.SQL, "LIMIT", sql+" LIMIT", 1)
		case strings.Contains(q.SQL, "OFFSET"):
			q.SQL = strings.Replace(q.SQL, "OFFSET", sql+" OFFSET", 1)
		default:
			q.SQL += " " + sql
		}
	}

	return nil
}

// Parse builds a SQL query from the supplied base query and URL values.
func (q *Query) Parse() error {
	if q.setStart < 1 {
		return errors.New(errors.ErrInvalidRequest,
			"invalid number of parameters passed to query")
	}

	order, groupBy := "", ""

	offset := " OFFSET 0"

	q.SQL = q.Base

	if q.Search != nil && q.Type != QueryInsert {
		if q.Search.Search != "" {
			if err := q.parseSearch(); err != nil {
				q.SQL = ""

				return err
			}
		}

		if q.Search.From != 0 {
			offset = fmt.Sprintf(" OFFSET %d", q.Search.From)
		}

		if q.Search.Summary != "" {
			s := strings.Split(q.Search.Summary, ",")

			for i, sv := range s {
				qf := q.Field(sv)
				if qf == nil {
					return errors.New(errors.ErrInvalidRequest,
						"invalid query summary value: "+sv)
				}

				if i == 0 {
					groupBy = " GROUP BY"
				} else {
					groupBy += ","
				}

				if qf.Table == "" {
					groupBy += " " + qf.Name
				} else {
					groupBy += " " + qf.Table + "." + qf.Name
				}
			}
		}

		if q.Search.Order != "" {
			s := strings.Split(q.Search.Order, ",")

			for i, sv := range s {
				dir := " ASC"

				if strings.HasPrefix(sv, "-") {
					sv = sv[1:]
					dir = " DESC"
				}

				qf := q.Field(sv)
				if qf == nil {
					return errors.New(errors.ErrInvalidRequest,
						"invalid query order value: "+sv)
				}

				if i == 0 {
					order = " ORDER BY"
				} else {
					order += ","
				}

				if qf.Table == "" {
					order += " " + qf.Name + dir
				} else {
					order += " " + qf.Table + "." + qf.Name + dir
				}
			}
		}
	}

	const userIDQuery = "(SELECT user_key FROM \"user\" WHERE user_id = $%d)"

	switch q.Type {
	case QuerySelect:
		if q.Search != nil {
			q.Limit = q.Search.Size
		}

		if q.Limit == 0 {
			q.Limit = q.Config.DBDefaultSize()
		} else if q.Limit < 0 || q.Limit > q.Config.DBMaxSize() {
			return errors.New(errors.ErrInvalidRequest,
				fmt.Sprintf("invalid query size value: %d "+
					"(must be between 1 and %d)",
					q.Limit, q.Config.DBMaxSize()))
		}

		if groupBy != "" && !strings.Contains(q.Base, "GROUP BY") {
			q.SQL += groupBy
		}

		if order != "" && !strings.Contains(q.Base, "ORDER BY") {
			q.SQL += order
		}

		if !strings.Contains(q.Base, "LIMIT") {
			if q.Limit > 1 {
				// Fetch one more than limit rows to test for more results.
				q.SQL += fmt.Sprintf(" LIMIT %d", q.Limit+1)
			} else {
				q.SQL += " LIMIT 1"
			}
		}

		if offset != "" && !strings.Contains(q.Base, "OFFSET") {
			q.SQL += offset
		}
	case QueryUpdate:
		sets := ""

		for i, sf := range q.Sets {
			if i > 0 {
				sets += ", "
			}

			f := q.Field(sf)

			switch {
			case f != nil && f.Table != q.Fields[0].Table &&
				strings.HasSuffix(sf, "_by"):
				sets += fmt.Sprintf("%s = "+userIDQuery,
					sf, q.setStart+int64(i))
			case f != nil && f.Table != q.Fields[0].Table && f.Key != "":
				from := f.Table

				if f.From != "" {
					from = f.From
				}

				join := sf

				if f.Join != "" {
					join = f.Join
				}

				sets += fmt.Sprintf("%s = "+
					"(SELECT %s FROM %s WHERE %s = $%d)",
					f.Key, f.Key, from, join, q.setStart+int64(i))
			case f != nil && f.Type == FieldTime:
				sets += fmt.Sprintf("%s = to_timestamp($%d)",
					sf, q.setStart+int64(i))
			default:
				sets += fmt.Sprintf("%s = $%d", sf, q.setStart+int64(i))
			}
		}

		if li := strings.LastIndex(q.SQL, "SET"); li >= 0 {
			newSQL := q.SQL[:li] + "SET " + sets

			if li < len(q.SQL)-3 {
				newSQL += q.SQL[li+3:]
			}

			q.SQL = newSQL
		}
	case QueryInsert:
		setFields := ""

		setValues := ""

		for i, sf := range q.Sets {
			if i > 0 {
				setFields += ", "
				setValues += ", "
			}

			f := q.Field(sf)

			switch {
			case f != nil && f.Table != q.Fields[0].Table &&
				strings.HasSuffix(sf, "_by"):
				setFields += sf
				setValues += fmt.Sprintf(userIDQuery, q.setStart+int64(i))
			case f != nil && f.Table != q.Fields[0].Table && f.Key != "":
				from := f.Table

				if f.From != "" {
					from = f.From
				}

				join := sf

				if f.Join != "" {
					join = f.Join
				}

				setFields += f.Key
				setValues += fmt.Sprintf("(SELECT %s FROM %s "+
					"WHERE %s = $%d)",
					f.Key, from, join, q.setStart+int64(i))
			case f != nil && f.Type == FieldTime:
				setFields += sf
				setValues += "to_timestamp($" +
					strconv.FormatInt(q.setStart+int64(i), 10) + ")"
			default:
				setFields += sf
				setValues += "$" + strconv.FormatInt(q.setStart+int64(i), 10)
			}
		}

		q.SQL = strings.Replace(q.SQL, "() VALUES ()",
			"("+setFields+") VALUES ("+setValues+")", 1)

		if strings.Contains(q.SQL, "DO UPDATE SET") {
			sets := ""

			for i, sf := range q.Sets {
				if i > 0 {
					sets += ", "
				}

				f := q.Field(sf)

				switch {
				case f != nil && f.Table != q.Fields[0].Table &&
					strings.HasSuffix(sf, "_by"):
					sets += fmt.Sprintf("%s = "+userIDQuery, sf,
						q.setStart+int64(i))
				case f != nil && f.Table != q.Fields[0].Table && f.Key != "":
					from := f.Table

					if f.From != "" {
						from = f.From
					}

					join := sf

					if f.Join != "" {
						join = f.Join
					}

					sets += fmt.Sprintf("%s = "+
						"(SELECT %s FROM %s WHERE %s = $%d)",
						f.Key, f.Key, from, join, q.setStart+int64(i))
				case f != nil && f.Type == FieldTime:
					sets += fmt.Sprintf("%s = to_timestamp($%d)",
						sf, q.setStart+int64(i))
				default:
					sets += fmt.Sprintf("%s = $%d", sf, q.setStart+int64(i))
				}
			}

			q.SQL = strings.Replace(q.SQL, "DO UPDATE SET",
				"DO UPDATE SET "+sets, 1)
		}
	case QueryDelete:
		// Prevent accidental execution of any delete all query.
		if len(q.Params) == 0 || !strings.Contains(q.SQL, "WHERE ") {
			q.SQL = ""

			return errors.New(errors.ErrInvalidRequest,
				"invalid delete query without parameters")
		}
	}

	return nil
}

// Exec executes a SQL statement that does not return rows.
func (q *Query) Exec(ctx context.Context) (SQLResult, error) {
	if q.SQL == "" {
		if err := q.Parse(); err != nil {
			return nil, err
		}
	}

	if q.Tx != nil {
		return q.Tx.Exec(ctx, q.SQL, q.Params...)
	}

	return q.DB.Exec(ctx, q.SQL, q.Params...)
}

// Query executes the query and returns the sql rows.
func (q *Query) Query(ctx context.Context) (SQLRows, error) {
	if q.SQL == "" {
		if err := q.Parse(); err != nil {
			return nil, err
		}
	}

	if q.Tx != nil {
		return q.Tx.Query(ctx, q.SQL, q.Params...)
	}

	return q.DB.Query(ctx, q.SQL, q.Params...)
}

// QueryRow executes the query and returns a single row.
func (q *Query) QueryRow(ctx context.Context) (SQLRow, error) {
	if q.SQL == "" {
		if err := q.Parse(); err != nil {
			return nil, err
		}
	}

	if q.Tx != nil {
		return q.Tx.QueryRow(ctx, q.SQL, q.Params...), nil
	}

	return q.DB.QueryRow(ctx, q.SQL, q.Params...), nil
}
