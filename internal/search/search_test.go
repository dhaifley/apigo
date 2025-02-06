package search_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dhaifley/apid/internal/search"
)

func TestScanner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		tok   search.ScanToken
		lit   string
		num   int
	}{
		{
			input: "and(",
			tok:   search.TokenKeyword,
			lit:   "and",
			num:   1,
		},
		{
			input: "gt(",
			tok:   search.TokenKeyword,
			lit:   "gt",
			num:   1,
		},
		{
			input: "lt(",
			tok:   search.TokenKeyword,
			lit:   "lt",
			num:   1,
		},
		{
			input: "gte(",
			tok:   search.TokenKeyword,
			lit:   "gte",
			num:   1,
		},
		{
			input: "lte(",
			tok:   search.TokenKeyword,
			lit:   "lte",
			num:   1,
		},
		{
			input: "match(",
			tok:   search.TokenKeyword,
			lit:   "match",
			num:   1,
		},
		{
			input: "b\"dGVzdA==\"",
			tok:   search.TokenTagVal,
			lit:   "test",
			num:   1,
		},
		{
			input: "b[test]\"dGVzdA==\"",
			tok:   search.TokenTagVal,
			lit:   "test",
			num:   1,
		},
		{
			input: "b/dGVzdA==/",
			tok:   search.TokenTagVal,
			lit:   "/test/",
			num:   1,
		},
		{
			input: "and(b\"dGVzdA==\":test)",
			tok:   search.TokenTagCat,
			lit:   "test",
			num:   3,
		},
		{
			input: "and(b/dGVzdA==/:test)",
			tok:   search.TokenTagCat,
			lit:   "/test/",
			num:   3,
		},
		{
			input: "and(test)",
			tok:   search.TokenTagCat,
			lit:   "test",
			num:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			buf := bytes.NewBufferString(tt.input)

			s := search.NewScanner(buf)

			var err error

			var tok search.ScanToken

			lit := ""

			for i := 0; i < tt.num; i++ {
				tok, lit, err = s.Scan()
				if err != nil {
					t.Fatal(err)
				}
			}

			if tok != tt.tok {
				t.Errorf("Expected token type: %v, got: %v", tt.tok, tok)
			}

			if lit != tt.lit {
				t.Errorf("Expected literal %v, got: %v", tt.lit, lit)
			}
		})
	}
}

func TestParser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input  string
		decode bool
		eval   func(node *search.QueryNode) (bool, error)
		res    func(ast *search.QueryTree)
	}{
		{
			input: "test and(and(one,and(apple,core)),two)",
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Cat == "apple" || node.Cat == "id" ||
					node.Cat == "core" || node.Cat == "one" ||
					node.Cat == "two" {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Cat != "id" {
					t.Errorf("result does not match expected: id != %s",
						ast.Root.Nodes[0].Cat)
				}

				if ast.Root.Nodes[0].Val != "test" {
					t.Errorf("result does not match expected: test != %s",
						ast.Root.Nodes[0].Val)
				}

				if ast.Root.Nodes[1].Nodes[0].Nodes[0].Cat != "one" {
					t.Errorf("result does not match expected: one != %s",
						ast.Root.Nodes[1].Nodes[0].Nodes[0].Cat)
				}

				if ast.Root.Nodes[1].Nodes[0].Nodes[0].Val != "" {
					t.Errorf("result does not match expected:  != %s",
						ast.Root.Nodes[1].Nodes[0].Nodes[0].Val)
				}

				if ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[0].Cat != "apple" {
					t.Errorf("result does not match expected: apple != %s",
						ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[0].Cat)
				}

				if ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[0].Val != "" {
					t.Errorf("result does not match expected:  != %s",
						ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[0].Val)
				}

				if ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[1].Cat != "core" {
					t.Errorf("result does not match expected: core != %s",
						ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[1].Cat)
				}

				if ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[1].Val != "" {
					t.Errorf("result does not match expected:  != %s",
						ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[1].Val)
				}

				if ast.Root.Nodes[1].Nodes[1].Cat != "two" {
					t.Errorf("result does not match expected: two != %s",
						ast.Root.Nodes[1].Nodes[1].Cat)
				}

				if ast.Root.Nodes[1].Nodes[1].Val != "" {
					t.Errorf("result does not match expected:  != %s",
						ast.Root.Nodes[1].Nodes[1].Val)
				}
			},
		},
		{
			input: "test and(and(one,and(apple,core:*)),two:*)",
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Cat == "apple" || node.Cat == "id" ||
					node.Cat == "core" || node.Cat == "one" ||
					node.Cat == "two" {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Cat != "id" {
					t.Errorf("result does not match expected: id != %s",
						ast.Root.Nodes[0].Cat)
				}

				if ast.Root.Nodes[0].Val != "test" {
					t.Errorf("result does not match expected: test != %s",
						ast.Root.Nodes[0].Val)
				}

				if ast.Root.Nodes[1].Nodes[0].Nodes[0].Cat != "one" {
					t.Errorf("result does not match expected: one != %s",
						ast.Root.Nodes[1].Nodes[0].Nodes[0].Cat)
				}

				if ast.Root.Nodes[1].Nodes[0].Nodes[0].Val != "" {
					t.Errorf("result does not match expected:  != %s",
						ast.Root.Nodes[1].Nodes[0].Nodes[0].Val)
				}

				if ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[0].Cat != "apple" {
					t.Errorf("result does not match expected: apple != %s",
						ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[0].Cat)
				}

				if ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[0].Val != "" {
					t.Errorf("result does not match expected:  != %s",
						ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[0].Val)
				}

				if ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[1].Cat != "core" {
					t.Errorf("result does not match expected: core != %s",
						ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[1].Cat)
				}

				if ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[1].Val != "*" {
					t.Errorf("result does not match expected: * != %s",
						ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[1].Val)
				}

				if ast.Root.Nodes[1].Nodes[1].Cat != "two" {
					t.Errorf("result does not match expected: two != %s",
						ast.Root.Nodes[1].Nodes[1].Cat)
				}

				if ast.Root.Nodes[1].Nodes[1].Val != "*" {
					t.Errorf("result does not match expected: * != %s",
						ast.Root.Nodes[1].Nodes[1].Val)
				}
			},
		},
		{
			decode: true,
			input:  `and(test:*b"Lw=="*)`,
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Val == `*/*` {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Nodes[0].Val != `*/*` {
					t.Errorf("result does not match expected: %s",
						ast.Root.Nodes[0].Nodes[0].Val)
				}
			},
		},
		{
			decode: false,
			input:  `and(test:*b"Lw=="*)`,
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Val == `*b"Lw=="*` {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Nodes[0].Val != `*b"Lw=="*` {
					t.Errorf("result does not match expected: %s",
						ast.Root.Nodes[0].Nodes[0].Val)
				}
			},
		},
		{
			input: `and(test:/)`,
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Val == `/` {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Nodes[0].Val != `/` {
					t.Errorf("result does not match expected: %s",
						ast.Root.Nodes[0].Nodes[0].Val)
				}
			},
		},
		{
			input: `output[1]`,
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Val == `output[1]` {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Val !=
					`output[1]` {
					t.Errorf("result does not match expected: %s",
						ast.Root.Nodes[0].Val)
				}
			},
		},
		{
			input: `and(uuid:11223344-5566-7788-9900-aaccddeeffbb)`,
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Val == `11223344-5566-7788-9900-aaccddeeffbb` {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Nodes[0].Val !=
					`11223344-5566-7788-9900-aaccddeeffbb` {
					t.Errorf("result does not match expected: %s",
						ast.Root.Nodes[0].Nodes[0].Val)
				}
			},
		},
		{
			input: `and(id:"(f"*)`,
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Val == `(f*` {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Nodes[0].Val != `(f*` {
					t.Errorf("result does not match expected: %s",
						ast.Root.Nodes[0].Nodes[0].Val)
				}
			},
		},
		{
			input: "*",
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Val == "*" {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Val != "*" {
					t.Errorf("result does not match expected: %s",
						ast.Root.Nodes[0].Val)
				}
			},
		},
		{
			input: `"trans&parse_only=true`,
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Val == `trans&parse_only=true` {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Val != `trans&parse_only=true` {
					t.Errorf("result does not match expected: %s",
						ast.Root.Nodes[0].Val)
				}
			},
		},
		{
			input: "foo - bar baz and(foo:bar)",
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Cat == "foo" || node.Cat == "id" {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				exp := "foo - bar baz"
				if ast.Root.Nodes[0].Val != exp {
					t.Errorf("result does not match expected: %s != %s", exp,
						ast.Root.Nodes[0].Val)
				}
			},
		},
		{
			input: "foo and(and(foo:bar,or(bar:baz,foo:baz)),not(blah:1))",
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Cat == "foo" || node.Cat == "id" {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[1].Val != "baz" {
					t.Errorf("result does not match expected: baz != %s",
						ast.Root.Nodes[1].Nodes[0].Nodes[1].Nodes[1].Val)
				}

				if ast.Root.Nodes[1].Nodes[1].Nodes[0].Val != "1" {
					t.Errorf("result does not match expected: 1 != %s",
						ast.Root.Nodes[1].Nodes[1].Nodes[1].Val)
				}
			},
		},
		{
			input: "and(foo:bar,or(bar:baz,foo:baz),not(blah:1))",
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Cat == "foo" || node.Cat == "id" {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Nodes[1].Nodes[1].Val != "baz" {
					t.Errorf("result does not match expected: baz != %s",
						ast.Root.Nodes[1].Nodes[1].Nodes[1].Val)
				}
			},
		},
		{
			input: "and(id:test)",
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Cat == "id" {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Nodes[0].Val != "test" {
					t.Errorf("result does not match expected: test != %s",
						ast.Root.Nodes[0].Nodes[0].Val)
				}
			},
		},
		{
			input: "test and(apple)",
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Cat == "apple" || node.Cat == "id" {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Cat != "id" {
					t.Errorf("result does not match expected: id != %s",
						ast.Root.Nodes[0].Cat)
				}

				if ast.Root.Nodes[0].Val != "test" {
					t.Errorf("result does not match expected: test != %s",
						ast.Root.Nodes[0].Val)
				}

				if ast.Root.Nodes[1].Nodes[0].Cat != "apple" {
					t.Errorf("result does not match expected: apple != %s",
						ast.Root.Nodes[1].Nodes[0].Cat)
				}

				if ast.Root.Nodes[1].Nodes[0].Val != "" {
					t.Errorf("result does not match expected:  != %s",
						ast.Root.Nodes[1].Nodes[0].Val)
				}
			},
		},
		{
			input: "test and(apple,core)",
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Cat == "apple" || node.Cat == "id" ||
					node.Cat == "core" {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Cat != "id" {
					t.Errorf("result does not match expected: id != %s",
						ast.Root.Nodes[0].Cat)
				}

				if ast.Root.Nodes[0].Val != "test" {
					t.Errorf("result does not match expected: test != %s",
						ast.Root.Nodes[0].Val)
				}

				if ast.Root.Nodes[1].Nodes[0].Cat != "apple" {
					t.Errorf("result does not match expected: apple != %s",
						ast.Root.Nodes[1].Nodes[0].Cat)
				}

				if ast.Root.Nodes[1].Nodes[0].Val != "" {
					t.Errorf("result does not match expected:  != %s",
						ast.Root.Nodes[1].Nodes[0].Val)
				}

				if ast.Root.Nodes[1].Nodes[1].Cat != "core" {
					t.Errorf("result does not match expected: core != %s",
						ast.Root.Nodes[1].Nodes[1].Cat)
				}

				if ast.Root.Nodes[1].Nodes[1].Val != "" {
					t.Errorf("result does not match expected:  != %s",
						ast.Root.Nodes[1].Nodes[1].Val)
				}
			},
		},
		{
			input: "test",
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Cat == "id" {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Cat != "id" {
					t.Errorf("result does not match expected: id != %s",
						ast.Root.Nodes[0].Val)
				}

				if ast.Root.Nodes[0].Val != "test" {
					t.Errorf("result does not match expected: test != %s",
						ast.Root.Nodes[0].Val)
				}
			},
		},
		{
			input: "and(id:test1,not(id:test2))",
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Cat == "id" && node.Val == "test1" {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Nodes[1].Op != search.OpNot {
					t.Errorf("Expected operator: not, got: %v",
						ast.Root.Nodes[0].Op)
				}

				if ast.Root.Nodes[0].Nodes[1].Nodes[0].Val != "test2" {
					t.Errorf("Expected node value: test, got: %v",
						ast.Root.Nodes[0].Nodes[1].Nodes[0].Val)
				}
			},
		},
		{
			input: "and(and(id:*dis*,not(id:*dm*)),not(id:*`md*`))",
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Cat == "id" && node.Val == "*dis*" {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Nodes[1].Op != search.OpNot {
					t.Errorf("Expected operator: not, got: %v",
						ast.Root.Nodes[0].Nodes[1].Op)
				}

				if ast.Root.Nodes[0].Nodes[1].Nodes[0].Val != "*`md*`" {
					t.Errorf("Expected node value: *`md*`, got: %v",
						ast.Root.Nodes[0].Nodes[1].Nodes[0].Val)
				}
			},
		},
		{
			input: "and(apple)",
			eval: func(node *search.QueryNode) (bool, error) {
				if node.Cat == "apple" && node.Val == "" {
					return true, nil
				}

				return false, nil
			},
			res: func(ast *search.QueryTree) {
				if ast.Root.Nodes[0].Nodes[0].Cat != "apple" {
					t.Errorf("Expected node category: apple, got: %v",
						ast.Root.Nodes[0].Nodes[0].Cat)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			buf := bytes.NewBufferString(tt.input)

			p := search.NewParser(buf)

			p.DecodeBase64(tt.decode)

			ast, err := p.Parse()
			if err != nil {
				t.Fatal(err)
			}

			tt.res(ast)

			res, err := ast.Eval(tt.eval)
			if err != nil {
				t.Fatal(err)
			}

			if res != true {
				t.Error("Expected eval result: true, got: false")
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBufferString(`and(id: foo)`)

	p := search.NewParser(buf)

	_, err := p.Parse()
	if err == nil {
		t.Fatal("Expecting error got nil")
	}

	if !strings.Contains(err.Error(), "invalid whitespace") {
		t.Fatalf("Expecting whitespace error, got: %v", err.Error())
	}
}
