package mcp

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// This file implements a small self-contained expression parser and
// compiler that turns an equation string like "c = sqrt(a^2 + b^2)" into a
// polyform node subgraph, so callers don't have to hand-wire arithmetic one
// math node at a time. It only targets the node types that actually exist
// in github.com/EliCDavis/polyform/math (see equationNodeTypes below) —
// notably there is no general pow(base, exponent), sin/cos/tan, or abs node
// in polyform today, so those aren't supported; see compileCall/compilePow.

// ============================================================================
// Lexer
// ============================================================================

type eqTokenKind int

const (
	eqTokNum eqTokenKind = iota
	eqTokIdent
	eqTokPlus
	eqTokMinus
	eqTokStar
	eqTokSlash
	eqTokCaret
	eqTokLParen
	eqTokRParen
	eqTokComma
	eqTokEOF
)

type eqToken struct {
	kind eqTokenKind
	text string
	num  float64
}

type eqLexer struct {
	input []rune
	pos   int
}

func newEqLexer(s string) *eqLexer {
	return &eqLexer{input: []rune(s)}
}

func (l *eqLexer) next() (eqToken, error) {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
	if l.pos >= len(l.input) {
		return eqToken{kind: eqTokEOF}, nil
	}

	c := l.input[l.pos]
	switch c {
	case '+':
		l.pos++
		return eqToken{kind: eqTokPlus}, nil
	case '-':
		l.pos++
		return eqToken{kind: eqTokMinus}, nil
	case '*':
		l.pos++
		return eqToken{kind: eqTokStar}, nil
	case '/':
		l.pos++
		return eqToken{kind: eqTokSlash}, nil
	case '^':
		l.pos++
		return eqToken{kind: eqTokCaret}, nil
	case '(':
		l.pos++
		return eqToken{kind: eqTokLParen}, nil
	case ')':
		l.pos++
		return eqToken{kind: eqTokRParen}, nil
	case ',':
		l.pos++
		return eqToken{kind: eqTokComma}, nil
	}

	if unicode.IsDigit(c) || c == '.' {
		start := l.pos
		seenDot := false
		for l.pos < len(l.input) && (unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '.') {
			if l.input[l.pos] == '.' {
				if seenDot {
					break
				}
				seenDot = true
			}
			l.pos++
		}
		text := string(l.input[start:l.pos])
		v, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return eqToken{}, fmt.Errorf("invalid number %q", text)
		}
		return eqToken{kind: eqTokNum, num: v, text: text}, nil
	}

	if unicode.IsLetter(c) || c == '_' {
		start := l.pos
		for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
			l.pos++
		}
		return eqToken{kind: eqTokIdent, text: string(l.input[start:l.pos])}, nil
	}

	return eqToken{}, fmt.Errorf("unexpected character %q", string(c))
}

// ============================================================================
// Parser
// ============================================================================

type eqExprKind int

const (
	eqExprNum eqExprKind = iota
	eqExprVar
	eqExprBinOp
	eqExprNeg
	eqExprCall
)

type eqExpr struct {
	kind eqExprKind
	num  float64
	name string // variable name, or function name for calls
	op   byte   // '+' '-' '*' '/' '^'
	args []*eqExpr
}

type eqParser struct {
	lex *eqLexer
	cur eqToken
}

func newEqParser(s string) (*eqParser, error) {
	p := &eqParser{lex: newEqLexer(s)}
	return p, p.advance()
}

func (p *eqParser) advance() error {
	t, err := p.lex.next()
	if err != nil {
		return err
	}
	p.cur = t
	return nil
}

func (p *eqParser) parseExpr() (*eqExpr, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for p.cur.kind == eqTokPlus || p.cur.kind == eqTokMinus {
		op := byte('+')
		if p.cur.kind == eqTokMinus {
			op = '-'
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &eqExpr{kind: eqExprBinOp, op: op, args: []*eqExpr{left, right}}
	}
	return left, nil
}

func (p *eqParser) parseTerm() (*eqExpr, error) {
	left, err := p.parsePower()
	if err != nil {
		return nil, err
	}
	for p.cur.kind == eqTokStar || p.cur.kind == eqTokSlash {
		op := byte('*')
		if p.cur.kind == eqTokSlash {
			op = '/'
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.parsePower()
		if err != nil {
			return nil, err
		}
		left = &eqExpr{kind: eqExprBinOp, op: op, args: []*eqExpr{left, right}}
	}
	return left, nil
}

func (p *eqParser) parsePower() (*eqExpr, error) {
	base, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	if p.cur.kind == eqTokCaret {
		if err := p.advance(); err != nil {
			return nil, err
		}
		exp, err := p.parsePower() // right-associative
		if err != nil {
			return nil, err
		}
		return &eqExpr{kind: eqExprBinOp, op: '^', args: []*eqExpr{base, exp}}, nil
	}
	return base, nil
}

func (p *eqParser) parseUnary() (*eqExpr, error) {
	if p.cur.kind == eqTokMinus {
		if err := p.advance(); err != nil {
			return nil, err
		}
		inner, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		// Fold "-<literal>" directly into a negative numeric literal
		// rather than wrapping it in an eqExprNeg node. This matters
		// beyond node-count: it's what lets an exponent like x^-1 be
		// recognized as a numeric-literal exponent by
		// equationCompiler.power, which only accepts eqExprNum (there's
		// no general pow node to fall back on for a negation-wrapped one).
		if inner.kind == eqExprNum {
			return &eqExpr{kind: eqExprNum, num: -inner.num}, nil
		}
		return &eqExpr{kind: eqExprNeg, args: []*eqExpr{inner}}, nil
	}
	return p.parseAtom()
}

func (p *eqParser) parseAtom() (*eqExpr, error) {
	switch p.cur.kind {
	case eqTokNum:
		v := p.cur.num
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &eqExpr{kind: eqExprNum, num: v}, nil

	case eqTokIdent:
		name := p.cur.text
		if err := p.advance(); err != nil {
			return nil, err
		}
		if p.cur.kind == eqTokLParen {
			if err := p.advance(); err != nil {
				return nil, err
			}
			var args []*eqExpr
			if p.cur.kind != eqTokRParen {
				for {
					a, err := p.parseExpr()
					if err != nil {
						return nil, err
					}
					args = append(args, a)
					if p.cur.kind == eqTokComma {
						if err := p.advance(); err != nil {
							return nil, err
						}
						continue
					}
					break
				}
			}
			if p.cur.kind != eqTokRParen {
				return nil, fmt.Errorf("expected ')' to close arguments to %s(...)", name)
			}
			if err := p.advance(); err != nil {
				return nil, err
			}
			return &eqExpr{kind: eqExprCall, name: name, args: args}, nil
		}
		return &eqExpr{kind: eqExprVar, name: name}, nil

	case eqTokLParen:
		if err := p.advance(); err != nil {
			return nil, err
		}
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.cur.kind != eqTokRParen {
			return nil, fmt.Errorf("expected ')'")
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
		return e, nil

	default:
		return nil, fmt.Errorf("unexpected token in expression")
	}
}

func isValidEqIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 && !(unicode.IsLetter(r) || r == '_') {
			return false
		}
		if i > 0 && !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			return false
		}
	}
	return true
}

// parseEquation splits "output = expression" and parses the right-hand
// side into an AST. The left-hand side must be a single identifier — it
// becomes the subgraph's boundary output name.
func parseEquation(s string) (outputName string, ast *eqExpr, err error) {
	idx := strings.Index(s, "=")
	if idx < 0 {
		return "", nil, fmt.Errorf("equation must contain '=', e.g. \"c = sqrt(a^2 + b^2)\"")
	}

	lhs := strings.TrimSpace(s[:idx])
	if !isValidEqIdent(lhs) {
		return "", nil, fmt.Errorf("left-hand side must be a single output name (letters/digits/underscore, not starting with a digit), got %q", lhs)
	}

	p, err := newEqParser(s[idx+1:])
	if err != nil {
		return "", nil, err
	}
	e, err := p.parseExpr()
	if err != nil {
		return "", nil, err
	}
	if p.cur.kind != eqTokEOF {
		return "", nil, fmt.Errorf("unexpected trailing input near %q", p.cur.text)
	}

	return lhs, e, nil
}
