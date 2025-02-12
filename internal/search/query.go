package search

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/dhaifley/apigo/internal/errors"
)

// ScanToken values represent individual tokens found by query scanners.
type ScanToken uint

// scanToken values used by the scanners to identify tokens.
const (
	TokenEOF ScanToken = iota
	TokenIllegal
	TokenWS // Whitespace
	TokenOP // Open parenthesis
	TokenCP // Close parenthesis
	TokenColon
	TokenComma
	TokenKeyword
	TokenTagCat
	TokenTagVal
)

// escapeWildcard escapes any special wildcard characters in a string.
func escapeWildcard(s string) string {
	s = strings.ReplaceAll(s, "?", "÷")
	s = strings.ReplaceAll(s, "*", "°")

	return s
}

// QueryOp values represent the possible operations for a query node.
type QueryOp string

// Query operation types.
const (
	OpMatch QueryOp = QueryOp("match")
	OpAnd   QueryOp = QueryOp("and")
	OpOr    QueryOp = QueryOp("or")
	OpNot   QueryOp = QueryOp("not")
	OpGT    QueryOp = QueryOp("gt")
	OpGTE   QueryOp = QueryOp("gte")
	OpLT    QueryOp = QueryOp("lt")
	OpLTE   QueryOp = QueryOp("lte")
)

// String returns the value of a query operator as a string.
func (qo QueryOp) String() string {
	return string(qo)
}

// QueryOpFromString returns a QueryOp value from its string name.
func QueryOpFromString(s string) QueryOp {
	for _, op := range []QueryOp{
		OpAnd,
		OpOr,
		OpNot,
		OpGT,
		OpGTE,
		OpLT,
		OpLTE,
	} {
		if strings.TrimSpace(strings.ToLower(s)) == op.String() {
			return op
		}
	}

	return OpMatch
}

// QueryNode values represent individual nodes in the query AST.
type QueryNode struct {
	Op    QueryOp      `json:"op,omitempty"`
	Comp  QueryOp      `json:"comp,omitempty"`
	Cat   string       `json:"cat,omitempty"`
	CatRE string       `json:"cat_re,omitempty"`
	Val   string       `json:"val,omitempty"`
	ValRE string       `json:"val_re,omitempty"`
	Nodes []*QueryNode `json:"args,omitempty"`
}

// NewQueryNode creates a new query node value based on the supplied parameters.
func NewQueryNode(op, comp QueryOp,
	cat, val string,
	nodes ...*QueryNode,
) *QueryNode {
	node := &QueryNode{Op: op, Comp: comp, Nodes: nodes}

	if strings.HasPrefix(cat, "/") && strings.HasSuffix(cat, "/") &&
		len(cat) > 2 {
		node.CatRE = strings.TrimSuffix(strings.TrimPrefix(cat, "/"), "/")
	} else {
		node.Cat = cat
	}

	if strings.HasPrefix(val, "/") && strings.HasSuffix(val, "/") &&
		len(val) > 2 {
		node.ValRE = strings.TrimSuffix(strings.TrimPrefix(val, "/"), "/")
	} else {
		node.Val = val
	}

	return node
}

// String returns a representation of the query node as a string.
func (qn *QueryNode) String() string {
	str, err := json.Marshal(qn)
	if err != nil {
		return ""
	}

	return string(str)
}

// Map returns a representation of the AST node in the format used by the UI.
func (qn *QueryNode) Map() map[string]any {
	v := map[string]any{}

	v["op"] = qn.Op.String()

	if v["op"] == OpMatch.String() {
		v["op"] = "literal"
	}

	v["comp"] = qn.Comp.String()

	if len(qn.Nodes) > 0 {
		for i, n := range qn.Nodes {
			v[fmt.Sprintf("param_%d", i)] = n.Map()
		}
	}

	if qn.Op == OpMatch {
		v["key"] = qn.Cat
		if qn.CatRE != "" {
			v["key"] = "/" + qn.CatRE + "/"
		}

		v["value"] = qn.Val
		if qn.ValRE != "" {
			v["value"] = "/" + qn.ValRE + "/"
		}
	}

	return v
}

// Eval steps through the AST and executes the provided function on each node
// and applies each query operation, returning a bool result.
func (qn *QueryNode) Eval(f func(node *QueryNode) (bool, error)) (bool, error) {
	var err error

	def := false

	res := false

	if qn.Op == OpMatch {
		return f(qn)
	}

	for _, n := range qn.Nodes {
		res, err = n.Eval(f)
		if err != nil {
			return false, err
		}

		switch qn.Op {
		case OpAnd, OpGT, OpGTE, OpLT, OpLTE, OpMatch:
			if !res {
				return false, nil
			}

			def = true
		case OpOr:
			if res {
				return true, nil
			}
		case OpNot:
			return !res, nil
		}
	}

	return def, nil
}

// QueryTree values represent abstract syntax trees for search queries.
type QueryTree struct {
	Root *QueryNode `json:"root,omitempty"`
}

// NewQueryTree creates and initializes a new metric name value.
func NewQueryTree() *QueryTree {
	return &QueryTree{Root: NewQueryNode(OpAnd, OpMatch, "", "")}
}

// String returns a representation of the query tree as a string.
func (qt *QueryTree) String() string {
	str, err := json.Marshal(qt)
	if err != nil {
		return ""
	}

	return string(str)
}

// Eval steps through the AST and executes the provided function on each node
// and applies each query operation, returning a bool result.
func (qt *QueryTree) Eval(f func(node *QueryNode) (bool, error)) (bool, error) {
	if qt.Root == nil {
		return false, nil
	}

	return qt.Root.Eval(f)
}

// Query messages represent query string search requests.
type Query struct {
	Search  string `json:"search,omitempty"`
	Size    int64  `json:"size,omitempty"`
	Skip    int64  `json:"skip,omitempty"`
	Sort    string `json:"sort,omitempty"`
	Summary string `json:"summary,omitempty"`
}

// NoSummary returns a copy of the query without the summary component.
func (q *Query) NoSummary() *Query {
	if q == nil {
		return nil
	}

	return &Query{
		Search: q.Search,
		Size:   q.Size,
		Skip:   q.Skip,
		Sort:   q.Sort,
	}
}

// ParseQuery parses a string in query string format into a Query value that
// can be used for search functions.
func ParseQuery(values url.Values) (*Query, error) {
	req := &Query{}

	for qk, qv := range values {
		qk = strings.ToLower(qk)

		if len(qv) == 0 {
			continue
		}

		switch qk {
		case "search":
			req.Search = qv[0]
		case "skip":
			if strings.TrimSpace(qv[0]) != "" {
				i, err := strconv.ParseInt(strings.TrimSpace(qv[0]), 10, 64)
				if err != nil || i < 0 {
					return nil, errors.New(errors.ErrInvalidRequest,
						"invalid query skip value",
						"query", values)
				}

				req.Skip = i
			}
		case "size":
			if strings.TrimSpace(qv[0]) != "" {
				i, err := strconv.ParseInt(strings.TrimSpace(qv[0]), 10, 64)
				if err != nil || i < 0 {
					return nil, errors.New(errors.ErrInvalidRequest,
						"invalid query size value",
						"query", values)
				}

				req.Size = i
			}
		case "sort":
			req.Sort = strings.Join(qv, ",")
		case "summary":
			req.Summary = strings.Join(qv, ",")
		}
	}

	return req, nil
}
