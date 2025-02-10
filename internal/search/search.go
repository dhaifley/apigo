// Package search is used to build or evaluate search queries.
package search

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"io"
	"strings"

	"github.com/dhaifley/apigo/internal/errors"
)

// Scanner represents a lexical scanner for search queries.
type Scanner struct {
	r      *bufio.Reader
	kw     bool
	name   bool
	cat    bool
	decB64 bool
}

// NewScanner returns a new Scanner value.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		r:      bufio.NewReader(r),
		name:   false,
		kw:     false,
		decB64: true,
	}
}

// read reads the next rune from the buffered reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (qs *Scanner) read() rune {
	ch, _, err := qs.r.ReadRune()
	if err != nil {
		return rune(0)
	}

	return ch
}

// unread places the previously read rune back on the buffered reader.
func (qs *Scanner) unread() error { return qs.r.UnreadRune() }

// Scan returns the next token and literal value.
func (qs *Scanner) Scan() (ScanToken, string, error) {
	ch := qs.read()

	if ch == rune(0) {
		return TokenEOF, "", nil
	}

	switch ch {
	case '(':
		return TokenOP, string(ch), nil
	case ')':
		return TokenCP, string(ch), nil
	case ':':
		return TokenColon, string(ch), nil
	case ',':
		qs.cat = true

		return TokenComma, string(ch), nil
	case ' ', '\t':
		return TokenWS, string(ch), nil
	}

	if err := qs.unread(); err != nil {
		return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
			"unable to unread to scan buffer")
	}

	tok, lit, err := qs.scanKeyword()
	if err != nil {
		return TokenIllegal, "", err
	}

	if tok != TokenIllegal {
		qs.kw = true
		qs.name = true
		qs.cat = true

		return tok, lit, nil
	}

	tok, lit, err = qs.scanTagToken()
	if err != nil {
		return TokenIllegal, "", err
	}

	if tok != TokenIllegal {
		return tok, lit, nil
	}

	return TokenIllegal, string(ch), nil
}

// scanKeyword attempts to read a keyword token from the scan buffer.
func (qs *Scanner) scanKeyword() (ScanToken, string, error) {
	var buf bytes.Buffer

	ch := qs.read()

	if ch == 'a' || ch == 'n' {
		if err := qs.unread(); err != nil {
			return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
				"unable to unread to scan buffer")
		}

		if chN, err := qs.r.Peek(4); err == nil &&
			(string(chN) == "and(" || string(chN) == "not(") {
			for i := 0; i < 3; i++ {
				_, err := buf.WriteRune(qs.read())
				if err != nil {
					return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
						"unable to write to token buffer")
				}
			}

			return TokenKeyword, buf.String(), nil
		}

		return TokenIllegal, "", nil
	} else if ch == 'o' {
		if err := qs.unread(); err != nil {
			return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
				"unable to unread to scan buffer")
		}

		if chN, err := qs.r.Peek(3); err == nil && string(chN) == "or(" {
			for i := 0; i < 2; i++ {
				_, err := buf.WriteRune(qs.read())
				if err != nil {
					return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
						"unable to write to token buffer")
				}
			}

			return TokenKeyword, buf.String(), nil
		}

		return TokenIllegal, "", nil
	} else if ch == 'g' || ch == 'l' {
		if err := qs.unread(); err != nil {
			return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
				"unable to unread to scan buffer")
		}

		if chN, err := qs.r.Peek(4); err == nil &&
			(string(chN) == "gte(" || string(chN) == "lte(") {
			for i := 0; i < 3; i++ {
				_, err := buf.WriteRune(qs.read())
				if err != nil {
					return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
						"unable to write to token buffer")
				}
			}

			return TokenKeyword, buf.String(), nil
		}

		if chN, err := qs.r.Peek(3); err == nil &&
			(string(chN) == "gt(" || string(chN) == "lt(") {
			for i := 0; i < 2; i++ {
				_, err := buf.WriteRune(qs.read())
				if err != nil {
					return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
						"unable to write to token buffer")
				}
			}

			return TokenKeyword, buf.String(), nil
		}

		return TokenIllegal, "", nil
	} else if ch == 'm' {
		if err := qs.unread(); err != nil {
			return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
				"unable to unread to scan buffer")
		}

		if chN, err := qs.r.Peek(6); err == nil && string(chN) == "match(" {
			for i := 0; i < 5; i++ {
				_, err := buf.WriteRune(qs.read())
				if err != nil {
					return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
						"unable to write to token buffer")
				}
			}

			return TokenKeyword, buf.String(), nil
		}

		return TokenIllegal, "", nil
	}

	if err := qs.unread(); err != nil {
		return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
			"unable to unread to scan buffer")
	}

	return TokenIllegal, "", nil
}

// scanQuoted scans until the end of a quoted section and returns the errors
// as a string.
func (qs *Scanner) scanQuoted(quote, end, escape rune) (string, error) {
	qBuf := &bytes.Buffer{}

	var ch, lastCh rune

	for {
		ch = qs.read()

		if escape != rune(0) && ch == escape {
			lastCh = ch

			ch = qs.read()

			if ch == quote {
				if _, err := qBuf.WriteRune(ch); err != nil {
					return "", errors.Wrap(err, errors.ErrSearch,
						"unable to write to quote buffer")
				}

				continue
			}

			if _, err := qBuf.WriteRune(lastCh); err != nil {
				return "", errors.Wrap(err, errors.ErrSearch,
					"unable to write to quote buffer")
			}
		}

		if ch == end || ch == rune(0) {
			break
		}

		if _, err := qBuf.WriteRune(ch); err != nil {
			return "", errors.Wrap(err, errors.ErrSearch,
				"unable to write to quote buffer")
		}
	}

	return qBuf.String(), nil
}

// scanTagToken attempts to read a tag token from the scan buffer.
func (qs *Scanner) scanTagToken() (ScanToken, string, error) {
	buf := &bytes.Buffer{}

	lastCh := rune(0)

loop:
	for {
		ch := qs.read()

		switch {
		case ch == '\\':
			lastCh = ch
			ch = qs.read()
			if ch == '"' {
				if _, err := buf.WriteRune(ch); err != nil {
					return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
						"unable to write to quote buffer")
				}

				lastCh = ch

				continue loop
			}

			if _, err := buf.WriteRune(lastCh); err != nil {
				return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
					"unable to write to tag token buffer")
			}

			if err := qs.unread(); err != nil {
				return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
					"unable to unread to scan buffer")
			}
		case ch == '"':
			qStr, err := qs.scanQuoted('"', '"', '\\')
			if err != nil {
				return TokenIllegal, "", err
			}

			if _, err := buf.WriteString(escapeWildcard(qStr)); err != nil {
				return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
					"unable to write to term buffer")
			}

			lastCh = '"'
		case ch == 'b':
			lastCh = ch

			ch = qs.read()

			if ch == '[' {
				if _, err := qs.scanQuoted(ch, ']', rune(0)); err != nil {
					return TokenIllegal, "", err
				}

				ch = qs.read()
			}

			if ch == '"' || ch == '/' {
				qStr, err := qs.scanQuoted(ch, ch, rune(0))
				if err != nil {
					return TokenIllegal, "", err
				}

				if !qs.decB64 {
					if _, err := buf.WriteRune('b'); err != nil {
						return TokenIllegal, "", errors.Wrap(err,
							errors.ErrSearch,
							"unable to write to tag token buffer")
					}

					if _, err := buf.WriteRune(ch); err != nil {
						return TokenIllegal, "", errors.Wrap(err,
							errors.ErrSearch,
							"unable to write to tag token buffer")
					}

					if _, err := buf.Write([]byte(qStr)); err != nil {
						return TokenIllegal, "", errors.Wrap(err,
							errors.ErrSearch,
							"unable to write to term buffer")
					}

					if _, err := buf.WriteRune(ch); err != nil {
						return TokenIllegal, "", errors.Wrap(err,
							errors.ErrSearch,
							"unable to write to tag token buffer")
					}

					lastCh = ch

					continue loop
				}

				dec, err := base64.StdEncoding.DecodeString(qStr)
				if err != nil {
					return TokenIllegal, "", errors.Wrap(err,
						errors.ErrInvalidRequest,
						"unable to decode base64 token")
				}

				if ch == '/' {
					if _, err := buf.WriteRune(ch); err != nil {
						return TokenIllegal, "", errors.Wrap(err,
							errors.ErrSearch,
							"unable to write to tag token buffer")
					}
				}

				if _, err := buf.Write(dec); err != nil {
					return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
						"unable to write to term buffer")
				}

				if ch == '/' {
					if _, err := buf.WriteRune(ch); err != nil {
						return TokenIllegal, "", errors.Wrap(err,
							errors.ErrSearch,
							"unable to write to tag token buffer")
					}
				}

				lastCh = ch

				continue loop
			}

			if _, err := buf.WriteRune(lastCh); err != nil {
				return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
					"unable to write to tag token buffer")
			}

			if err := qs.unread(); err != nil {
				return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
					"unable to unread to scan buffer")
			}

			lastCh = ch
		case qs.name && ch == ':':
			qs.name = false
			lastCh = ch

			break loop
		case qs.kw && (ch == ',' || ch == ')'):
			lastCh = ch
			qs.name = true

			if err := qs.unread(); err != nil {
				return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
					"unable to unread to scan buffer")
			}

			break loop
		case !qs.kw && ch == ' ':
			lastCh = ch

			if chN, err := qs.r.Peek(4); err == nil &&
				(string(chN) == "and(" || string(chN) == "not(" ||
					strings.HasPrefix(string(chN), "or(")) {
				qs.name = true

				break loop
			} else if _, err := buf.WriteRune(ch); err != nil {
				return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
					"unable to write to tag token buffer")
			}
		case ch == rune(0): // EOF
			break loop
		default:
			lastCh = ch

			if _, err := buf.WriteRune(ch); err != nil {
				return TokenIllegal, "", errors.Wrap(err, errors.ErrSearch,
					"unable to write to tag token buffer")
			}
		}
	}

	if lastCh == ':' || qs.cat {
		qs.cat = false

		return TokenTagCat, buf.String(), nil
	}

	qs.cat = true

	return TokenTagVal, buf.String(), nil
}

// Parser values are used to parse query AST values.
type Parser struct {
	s       *Scanner
	Primary string
	pri     bool
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{
		s:   NewScanner(r),
		pri: true,
	}
}

// DecodeBase64 instructs the parser whether to decode or pass through base64
// encoded search terms.
func (qp *Parser) DecodeBase64(decode bool) {
	qp.s.decB64 = decode
}

// parse recursively scans and parses a query.
func (qp *Parser) parse(node *QueryNode) error {
	tok, lit, err := qp.s.Scan()
	if err != nil {
		return errors.Wrap(err, errors.ErrSearch,
			"unable to scan next token from query")
	}

	switch tok {
	case TokenEOF:
		return nil
	case TokenWS:
		return qp.parse(node)
	case TokenComma:
		return qp.parse(node)
	case TokenKeyword:
		if t, l, err := qp.s.Scan(); t != TokenOP {
			if err != nil {
				return errors.Wrap(err, errors.ErrInvalidRequest,
					"unable to parse keyword")
			}

			return errors.New(errors.ErrInvalidRequest,
				"parse failure, expecting ( got: "+l)
		}

		qp.pri = false

		newOp := QueryOpFromString(lit)

		newComp := newOp

		switch newOp {
		case OpAnd, OpOr, OpNot:
			newComp = node.Comp
		}

		newNode := NewQueryNode(newOp, newComp, "", "")

		if err := qp.parse(newNode); err != nil {
			return errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to parse child node")
		}

		if len(newNode.Nodes) == 0 {
			return errors.New(errors.ErrInvalidRequest,
				"unable to parse empty keyword")
		}

		node.Nodes = append(node.Nodes, newNode)

		t, l, err := qp.s.Scan()
		if err != nil {
			return errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to parse keyword")
		}

		if t != TokenCP && t != TokenComma && t != TokenEOF {
			return errors.New(errors.ErrInvalidRequest,
				"parse failure, expecting ) or , got: "+l)
		}

		if t == TokenCP || t == TokenEOF {
			return nil
		}

		return qp.parse(node)
	case TokenTagVal:
		// Only returns this for the primary search.
		newNode := NewQueryNode(OpMatch, node.Comp, "", lit)

		if qp.pri {
			newNode.Cat = "id"

			if qp.Primary != "" {
				newNode.Cat = qp.Primary
			}
		}

		node.Nodes = append(node.Nodes, newNode)

		return qp.parse(node)
	case TokenTagCat:
		t, l, err := qp.s.Scan()
		if err != nil {
			return errors.Wrap(err, errors.ErrSearch,
				"unable to parse tag value")
		}

		if t == TokenWS {
			return errors.Wrap(err, errors.ErrInvalidRequest,
				"invalid whitespace in query")
		}

		if t != TokenTagVal {
			l = ""
		}

		newNode := NewQueryNode(OpMatch, node.Comp, lit, l)

		node.Nodes = append(node.Nodes, newNode)

		if t == TokenTagVal {
			if t, l, err = qp.s.Scan(); err != nil {
				return errors.Wrap(err, errors.ErrInvalidRequest,
					"unable to parse keyword")
			}
		}

		if t != TokenCP && t != TokenComma && t != TokenEOF {
			return errors.New(errors.ErrInvalidRequest,
				"parse failure, expecting ) or , got: "+l)
		}

		if t == TokenCP || t == TokenEOF {
			return nil
		}

		return qp.parse(node)
	case TokenIllegal:
		return errors.New(errors.ErrInvalidRequest,
			"unable to parse query, illegal token: "+lit)
	default:
		return errors.New(errors.ErrInvalidRequest,
			"invalid  syntax query")
	}
}

// Parse scans and parses a query.
func (qp *Parser) Parse() (*QueryTree, error) {
	qt := NewQueryTree()

	if err := qp.parse(qt.Root); err != nil {
		return nil, err
	}

	return qt, nil
}
