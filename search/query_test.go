package search_test

import (
	"net/url"
	"testing"

	"github.com/dhaifley/apid/search"
)

func TestQueryString(t *testing.T) {
	t.Parallel()

	qt := search.NewQueryTree()
	qt.Root = &search.QueryNode{
		Op: search.OpAnd,
		Nodes: []*search.QueryNode{
			{
				Op:   search.OpMatch,
				Comp: search.OpGT,
				Cat:  "test",
				Val:  "test",
			},
		},
	}

	exp := `{"root":{"op":"and","args":[{"op":"match","comp":"gt",` +
		`"cat":"test","val":"test"}]}}`
	if qt.String() != exp {
		t.Errorf("Expected string: %v, got: %v", exp, qt.String())
	}
}

func TestQueryTreeMap(t *testing.T) {
	t.Parallel()

	qt := search.NewQueryTree()
	qt.Root = &search.QueryNode{
		Op:   search.OpAnd,
		Comp: search.OpMatch,
		Nodes: []*search.QueryNode{
			{
				Op:   search.OpMatch,
				Comp: search.OpMatch,
				Cat:  "__name",
				Val:  "test",
			},
			{
				Op:   search.OpOr,
				Comp: search.OpMatch,
				Nodes: []*search.QueryNode{
					{
						Op:   search.OpMatch,
						Comp: search.OpMatch,
						Cat:  "test",
						Val:  "test",
					},
					{
						Op:   search.OpMatch,
						Comp: search.OpMatch,
						Cat:  "test",
						Val:  "/test/",
					},
				},
			},
		},
	}
}

func TestQueryNoSummary(t *testing.T) {
	t.Parallel()

	q := &search.Query{
		Summary: "test",
	}

	res := q.NoSummary().Summary

	if res != "" {
		t.Errorf("Expected blank summary, got: %v", res)
	}
}

func TestParseQuery(t *testing.T) {
	t.Parallel()

	q := "search=test%20(test:test)&from=10&size=10&order=test" +
		"&ver=v2&search=(test1:test1)&order=-test1&summary=test,test1"

	values, err := url.ParseQuery(q)
	if err != nil {
		t.Fatal(err)
	}

	req, err := search.ParseQuery(values)
	if err != nil {
		t.Fatal(err)
	}

	expS := "test (test:test)"

	if req.Search != expS {
		t.Errorf("Expected search: %v, got: %v", expS, req.Search)
	}

	expI := int64(10)

	if req.Size != expI {
		t.Errorf("Expected size: %v, got: %v", expI, req.Size)
	}

	expI = int64(10)

	if req.From != expI {
		t.Errorf("Expected from: %v, got: %v", expI, req.From)
	}

	expS = "test,-test1"

	if req.Order != expS {
		t.Errorf("Expected order: %v, got: %v", expS, req.Order)
	}

	expS = "test,test1"

	if req.Summary != expS {
		t.Errorf("Expected summary: %v, got: %v", expS, req.Summary)
	}
}
