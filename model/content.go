package model

import (
	"fmt"
	"strconv"
	"strings"
)

// ContentMatch represents a match state of a node type's content expression,
// and can be used to find out whether further content matches here, and
// whether a given position is a valid end of the node.
type ContentMatch struct {
	// True when this match state represents a valid end of the node.
	ValidEnd  bool
	next      []interface{} // even indexes are *NodeType, odd are *ContentMatch
	wrapCache []interface{}
}

// NewContentMatch is the constructor for ContentMatch.
func NewContentMatch(validEnd bool) *ContentMatch {
	return &ContentMatch{
		ValidEnd:  validEnd,
		next:      nil,
		wrapCache: nil,
	}
}

func parseContentMatch(str string, nodeTypes map[string]*NodeType) (*ContentMatch, error) {
	stream := newTokenStream(str, nodeTypes)
	if stream.next() == nil {
		return EmptyContentMatch, nil
	}
	expr, err := parseExpr(stream)
	if err != nil {
		return nil, err
	}
	if stream.next() != nil {
		return nil, stream.err("Unexpected trailing text")
	}
	fmt.Printf("expr = %v\n", expr) // TODO
	return EmptyContentMatch, nil   // TODO
}

// MatchType matches a node type, returning a match after that node if
// successful.
func (cm *ContentMatch) MatchType(typ *NodeType) *ContentMatch {
	for i := 0; i < len(cm.next); i += 2 {
		if cm.next[i] == typ {
			return cm.next[i+1].(*ContentMatch)
		}
	}
	return nil
}

// MatchFragment tries to match a fragment. Returns the resulting match when
// successful.
//
// :: (Fragment, ?number, ?number) → ?ContentMatch
func (cm *ContentMatch) MatchFragment(frag *Fragment, args ...int) *ContentMatch {
	cur := cm
	start := 0
	if len(args) > 0 {
		start = args[0]
	}
	end := 0
	if len(args) > 1 {
		end = args[1]
	} else {
		end = frag.ChildCount()
	}
	for i := start; cur != nil && i < end; i++ {
		child, err := frag.Child(i)
		if err != nil {
			return nil
		}
		cur = cur.MatchType(child.Type)
	}
	return cur
}

func (cm *ContentMatch) inlineContent() bool {
	if len(cm.next) == 0 {
		return false
	}
	return cm.next[0].(*NodeType).IsInline()
}

func (cm *ContentMatch) compatible(other *ContentMatch) bool {
	for i := 0; i < len(cm.next); i += 2 {
		for j := 0; j < len(other.next); j += 2 {
			if cm.next[i] == other.next[j] {
				return true
			}
		}
	}
	return false
}

// EmptyContentMatch is an empty ContentMatch.
var EmptyContentMatch = NewContentMatch(true)

type tokenStream struct {
	str       string
	nodeTypes map[string]*NodeType
	inline    *bool
	pos       int
	tokens    []string
}

func newTokenStream(str string, nodeTypes map[string]*NodeType) *tokenStream {
	tokens := strings.Fields(str) // TODO string.split(/\s*(?=\b|\W|$)/)
	if tokens[len(tokens)-1] == "" {
		tokens = tokens[:len(tokens)-1]
	}
	if tokens[0] == "" {
		tokens = tokens[1:]
	}
	return &tokenStream{
		str:       str,
		nodeTypes: nodeTypes,
		tokens:    tokens,
	}
}

func (ts *tokenStream) next() *string {
	if ts.pos >= len(ts.tokens) {
		return nil
	}
	return &ts.tokens[ts.pos]
}

func (ts *tokenStream) eat(tok string) bool {
	if s := ts.next(); s != nil && *s != tok {
		return false
	}
	ts.pos++
	return true
}

func (ts *tokenStream) err(format string, args ...interface{}) error {
	str := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s (in content expression %q)", str, ts.str)
}

type exprType struct {
	Type  string
	Exprs []*exprType
	Expr  *exprType
	Min   int
	Max   int
	Value *NodeType
}

func parseExpr(stream *tokenStream) (*exprType, error) {
	exprs := []*exprType{}
	for {
		seq, err := parseExprSeq(stream)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, seq)
		if !stream.eat("|") {
			break
		}
	}
	if len(exprs) == 1 {
		return exprs[0], nil
	}
	return &exprType{Type: "choice", Exprs: exprs}, nil
}

func parseExprSeq(stream *tokenStream) (*exprType, error) {
	exprs := []*exprType{}
	for {
		sub, err := parseExprSubscript(stream)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, sub)
		if s := stream.next(); s != nil && *s != ")" && *s != "|" {
			break
		}
	}
	if len(exprs) == 1 {
		return exprs[0], nil
	}
	return &exprType{Type: "seq", Exprs: exprs}, nil
}

func parseExprSubscript(stream *tokenStream) (*exprType, error) {
	expr, err := parseExprAtom(stream)
	if err != nil {
		return nil, err
	}
	for {
		if stream.eat("+") {
			expr = &exprType{Type: "plus", Expr: expr}
		} else if stream.eat("*") {
			expr = &exprType{Type: "star", Expr: expr}
		} else if stream.eat("?") {
			expr = &exprType{Type: "opt", Expr: expr}
		} else if stream.eat("{") {
			expr, err = parseExprRange(stream, expr)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}
	return expr, nil
}

func parseNum(stream *tokenStream) (int, error) {
	s := stream.next()
	if s == nil {
		return 0, stream.err("Expected number, got nil")
	}
	result, err := strconv.Atoi(*s)
	if err != nil {
		return 0, stream.err("Expected number, got %q", *s)
	}
	return result, nil
}

func parseExprRange(stream *tokenStream, expr *exprType) (*exprType, error) {
	min, err := parseNum(stream)
	if err != nil {
		return nil, err
	}
	max := min
	if stream.eat(",") {
		if s := stream.next(); s != nil && *s != "}" {
			max, err = parseNum(stream)
			if err != nil {
				return nil, err
			}
		} else {
			max = -1
		}
	}
	if !stream.eat("}") {
		return nil, stream.err("Unclosed braced range")
	}
	return &exprType{Type: "range", Min: min, Max: max, Expr: expr}, nil
}

func resolveName(stream *tokenStream, name string) ([]*NodeType, error) {
	types := stream.nodeTypes
	if typ, ok := types[name]; ok {
		return []*NodeType{typ}, nil
	}
	var result []*NodeType
	for _, typ := range types {
		for _, g := range typ.Groups {
			if g == name {
				result = append(result, typ)
				break
			}
		}
	}
	if len(result) == 0 {
		return nil, stream.err("No node or type %q found", name)
	}
	return result, nil
}

func isWordCharacters(str string) bool {
	for _, c := range str {
		switch {
		case '0' <= c && c <= '9':
		case 'a' <= c && c <= 'z':
		case 'A' <= c && c <= 'Z':
		case c == '_':
			// OK
		default:
			return false
		}
	}
	return true
}

func parseExprAtom(stream *tokenStream) (*exprType, error) {
	if stream.eat("(") {
		expr, err := parseExpr(stream)
		if err != nil {
			return nil, err
		}
		if !stream.eat(")") {
			return nil, stream.err("Missing closing paren")
		}
		return expr, nil
	}

	s := stream.next()
	if s != nil && isWordCharacters(*s) {
		var exprs []*exprType
		s := stream.next()
		if s == nil {
			return nil, stream.err("Missing token")
		}
		types, err := resolveName(stream, *s)
		if err != nil {
			return nil, err
		}
		for _, typ := range types {
			inline := typ.IsInline()
			if stream.inline == nil {
				stream.inline = &inline
			} else if *stream.inline != inline {
				return nil, stream.err("Mixing inline and block content")
			}
			e := exprType{Type: "name", Value: typ}
			exprs = append(exprs, &e)
		}
		stream.pos++
		if len(exprs) == 1 {
			return exprs[0], nil
		}
		return &exprType{Type: "choice", Exprs: exprs}, nil
	}

	if s = stream.next(); s != nil {
		return nil, stream.err("Unexpected token %q", *s)
	}
	return nil, stream.err("Unexpected token nil")
}
