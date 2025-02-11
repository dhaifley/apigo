package sqldb

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/search"
)

// FieldType is an enum type describing the type of a field.
type FieldType string

// Supported search field type values.
const (
	FieldString = FieldType("string")
	FieldInt    = FieldType("int")
	FieldFloat  = FieldType("float")
	FieldBool   = FieldType("bool")
	FieldTime   = FieldType("time")
	FieldArray  = FieldType("array")
	FieldJSON   = FieldType("json")
)

// FieldOperator is an enum type describing the type of an operator.
type FieldOperator string

// Supported operator types.
const (
	OpEq    = FieldOperator("=")
	OpNotEq = FieldOperator("!=")
	OpGT    = FieldOperator(">")
	OpGTE   = FieldOperator(">=")
	OpLT    = FieldOperator("<")
	OpLTE   = FieldOperator("<=")
	OpAny   = FieldOperator("ANY")
	OpLike  = FieldOperator("LIKE")
	OpRE    = FieldOperator("~")
)

// FieldOption type values are used to specify options for field selection for
// queries.
type FieldOption string

// Supported field selection query options.
const (
	OptUserDetails = FieldOption("user_details")
)

// FieldOptions represent a collection of query options for field selection.
type FieldOptions []FieldOption

// Contains returns whether the collection contains a specified option.
func (fo *FieldOptions) Contains(option FieldOption) bool {
	if fo == nil {
		return false
	}

	for _, v := range *fo {
		if v == option {
			return true
		}
	}

	return false
}

// ParseFieldOptions parses options from query string values.
func ParseFieldOptions(values url.Values) (FieldOptions, error) {
	r := FieldOptions{}

	for qk, qv := range values {
		qk = strings.ToLower(qk)

		if len(qv) == 0 {
			continue
		}

		switch FieldOption(qk) {
		case OptUserDetails:
			b := strings.ToLower(strings.TrimSpace(qv[0]))
			if b != "0" && b != "f" && b != "false" {
				r = append(r, OptUserDetails)
			}
		}
	}

	return r, nil
}

// Field values represent individual query search fields.
type Field struct {
	Name     string        `json:"name,omitempty"`
	Type     FieldType     `json:"type,omitempty"`
	Table    string        `json:"table,omitempty"`
	From     string        `json:"from,omitempty"`
	Key      string        `json:"key,omitempty"`
	Join     string        `json:"join,omitempty"`
	JoinFrom string        `json:"join_from,omitempty"`
	Require  bool          `json:"require,omitempty"`
	Expr     string        `json:"expression,omitempty"`
	Op       FieldOperator `json:"operator,omitempty"`
	Option   FieldOption   `json:"option,omitempty"`
	Search   []string      `json:"search,omitempty"`
	Hidden   bool          `json:"hidden,omitempty"`
	Primary  bool          `json:"primary,omitempty"`
	Tags     bool          `json:"tags,omitempty"`
}

// String formats a field value as a JSON format string.
func (f *Field) String() string {
	str, err := json.Marshal(f)
	if err != nil {
		return ""
	}

	return string(str)
}

// SelectFields returns a SQL query SELECT stub for the specified fields.
func SelectFields(
	table string,
	fields []*Field,
	query *search.Query,
	options FieldOptions,
) string {
	res := "SELECT\n"

	joins, leftJoins, sumFields := []string{}, []string{}, []string{}

	first := false

	if query != nil && query.Summary != "" {
		sumFields = strings.Split(query.Summary, ",")
	}

	for _, f := range fields {
		if len(sumFields) > 0 {
			found := false

			for _, sf := range sumFields {
				if (f.Name == sf && f.Table == table) ||
					((f.Table + "." + f.Name) == sf) {
					found = true

					break
				}
			}

			if !found {
				continue
			}

			if first {
				res += ",\n"
			} else {
				first = true
			}

			res += "\t" + f.Table + "." + f.Name +
				" AS " + strings.Trim(f.Table, `"`) + "_" + f.Name

			continue
		}

		if f.Option != "" {
			found := false

			for _, o := range options {
				if f.Option == o {
					found = true

					break
				}
			}

			if !found {
				continue
			}
		}

		if !f.Hidden {
			if first {
				res += ",\n"
			} else {
				first = true
			}

			if f.Type == FieldTime {
				res += "\tEXTRACT(epoch FROM " + f.Table + "." + f.Name +
					")::BIGINT AS " + strings.Trim(f.Table, `"`) + "_" + f.Name
			} else if f.Tags {
				res += "\t" + `(SELECT
				ARRAY_AGG(tag_obj.tag_key || ':' || tag_obj.tag_val) AS tags
			FROM tag_obj
			WHERE tag_obj.status = '` + request.StatusActive + `'
				AND tag_obj.tag_type = '` + strings.Trim(f.Table, `"`) + `'
				AND tag_obj.tag_obj_id = ` + f.Table + `.` +
					strings.Trim(f.Table, `"`) + `_id::TEXT) AS ` + f.Name
			} else if f.Expr != "" {
				res += "\t" + f.Expr + " AS " + f.Name
			} else {
				res += "\t" + f.Table + "." + f.Name +
					" AS " + strings.Trim(f.Table, `"`) + "_" + f.Name
			}
		}

		if f.Table != table && (f.Key != "" || f.Join != "") {
			jq := "JOIN "

			if f.From != "" {
				jq += f.From + " "
			}

			key := f.Key

			if key == "" {
				key = f.Table + "_key"
			}

			jq += f.Table + " ON (" + f.Table + "." + key + " = "

			joinFrom := table

			if f.JoinFrom != "" {
				joinFrom = f.JoinFrom
			}

			jq += joinFrom + "."

			join := f.Join

			if join == "" {
				join = key
			}

			jq += join + ")"

			if f.Require {
				joins = append(joins, jq)
			} else {
				jq = "LEFT " + jq
				leftJoins = append(leftJoins, jq)
			}
		}
	}

	if len(sumFields) > 0 {
		res += ",\n\tCOUNT(*) AS count"
	}

	res += "\nFROM " + table

	for _, j := range joins {
		res += "\n" + j
	}

	for _, j := range leftJoins {
		res += "\n" + j
	}

	return res + "\n"
}

// SearchFields returns a SQL query SELECT stub for the specified table key
// field, joining for other fields as needed for search.
func SearchFields(
	table string,
	fields []*Field,
) string {
	res := "SELECT\n"

	gotKey, gotID := false, false

	joins, leftJoins := []string{}, []string{}

	for _, f := range fields {
		if !gotKey && f.Table == table && (strings.HasSuffix(f.Name, "_key") ||
			(table == "token" && f.Name == "token_id")) {
			res += "\t" + f.Table + "." + f.Name +
				" AS " + strings.Trim(f.Table, `"`) + "_" + f.Name

			gotKey = true
		}

		if gotKey && !gotID && f.Table == table && f.Table != "token" &&
			(strings.HasSuffix(f.Name, "_id")) {
			res += ",\n\t" + f.Table + "." + f.Name +
				" AS " + strings.Trim(f.Table, `"`) + "_" + f.Name

			gotID = true
		}

		if f.Table != table && (f.Key != "" || f.Join != "") {
			jq := "JOIN "

			if f.From != "" {
				jq += f.From + " "
			}

			key := f.Key

			if key == "" {
				key = f.Table + "_key"
			}

			jq += f.Table + " ON (" + f.Table + "." + key + " = "

			joinFrom := table

			if f.JoinFrom != "" {
				joinFrom = f.JoinFrom
			}

			jq += joinFrom + "."

			join := f.Join

			if join == "" {
				join = key
			}

			jq += join + ")"

			if f.Require {
				joins = append(joins, jq)
			} else {
				jq = "LEFT " + jq
				leftJoins = append(leftJoins, jq)
			}
		}
	}

	res += "\nFROM " + table

	for _, j := range joins {
		res += "\n" + j
	}

	for _, j := range leftJoins {
		res += "\n" + j
	}

	return res + "\n"
}

// ReturningFields returns a SQL query RETURNING clause for the specified
// fields.
func ReturningFields(
	table string,
	fields []*Field,
	options FieldOptions,
) string {
	if len(fields) == 0 {
		return ""
	}

	res := "\nRETURNING\n"

	first := false

	for i, f := range fields {
		if f.Hidden {
			continue
		}

		if f.Option != "" {
			found := false

			for _, o := range options {
				if f.Option == o {
					found = true

					break
				}
			}

			if !found {
				continue
			}
		}

		if first {
			res += ",\n"
		} else {
			first = true
		}

		if f.Expr != "" {
			res += "\t" + f.Expr + " AS " + f.Name
		} else if f.Table != table {
			alias := strings.Trim(f.Table, `"`) + "_" +
				strconv.FormatInt(int64(i), 10)

			key := f.Key

			if key == "" {
				key = f.Table + "_key"
			}

			join := f.Join

			if join == "" {
				join = key
			}

			res += "\t(SELECT " + alias + "." + f.Name +
				" AS " + alias + "_" + f.Name + " FROM "

			if f.From != "" {
				res += f.From + " "
			} else {
				res += f.Table + " "
			}

			res += alias + " WHERE " + alias + "." + key + " = " +
				table + "." + join + " LIMIT 1)"
		} else {
			if f.Type == FieldTime {
				res += "\tEXTRACT(epoch FROM " + f.Table + "." + f.Name +
					")::BIGINT AS " + strings.Trim(f.Table, `"`) + "_" + f.Name
			} else if f.Tags {
				res += "\t" + `(SELECT
				ARRAY_AGG(tag_obj.tag_key || ':' || tag_obj.tag_val) AS tags
			FROM tag_obj
			WHERE tag_obj.status = '` + request.StatusActive + `'
				AND tag_obj.tag_type = '` + strings.Trim(f.Table, `"`) + `'
				AND tag_obj.tag_obj_id = ` + f.Table + `.` +
					strings.Trim(f.Table, `"`) + `_id::TEXT) AS ` + f.Name
			} else {
				res += "\t" + f.Table + "." + f.Name +
					" AS " + strings.Trim(f.Table, `"`) + "_" + f.Name
			}
		}
	}

	return res + "\n"
}

// SummaryData values contain summary results.
type SummaryData map[string]any

// ScanDest returns the destination fields for a SQL row scan.
func (sd *SummaryData) ScanDest(
	fields []*Field,
	query *search.Query,
) []any {
	if sd == nil {
		return nil
	}

	res, sumFields := []any{}, []string{}

	if query != nil && query.Summary != "" {
		sumFields = strings.Split(query.Summary, ",")
	}

	for _, f := range fields {
		found := ""

		for _, sf := range sumFields {
			if (f.Name == sf && f.Table == fields[0].Table) ||
				((f.Table + "." + f.Name) == sf) {
				found = sf

				break
			}

			for _, ssf := range f.Search {
				if ssf == sf {
					found = sf

					break
				}
			}

			if found != "" {
				break
			}
		}

		if found == "" {
			continue
		}

		var v any

		switch f.Type {
		case FieldString:
			v = new(string)
		case FieldArray:
			vv := []any{}

			v = &vv
		case FieldBool:
			v = new(bool)
		case FieldFloat:
			v = new(float64)
		case FieldInt:
			v = new(int64)
		case FieldTime:
			v = new(int64)
		case FieldJSON:
			vv := map[string]any{}

			v = &vv
		}

		(*sd)[found] = v

		res = append(res, v)
	}

	cf := new(int64)

	(*sd)["count"] = cf

	res = append(res, cf)

	return res
}
