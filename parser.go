package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type SExpr interface{ sExpr() } // any type defined in package main that implements sExpr() with no args and no return will be considered to be of type SExpr

type Atom struct{ Value string } // an atom is just a string in Go ("a", "foo", "+", etc)

func (a Atom) sExpr() {} // tells Go that Atom should count as an sExpr

type Pair struct{ Car, Cdr SExpr }

func (p Pair) sExpr() {} // tells Go that Pair is an sExpr

type TokenType int // basically declaring an enum here using the iota keyword in Go

const (
	TokenEOF TokenType = iota
	TokenLParen
	TokenRParen
	TokenSymbol
	TokenQuote
)

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct{ reader *bufio.Reader }

func NewLexer(r io.Reader) *Lexer {
	return &Lexer{reader: bufio.NewReader(r)} // Go has escape analysis so this will actually be allocated on the heap not stack
}

func (l *Lexer) DiscardLine() {
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil || r == '\n' {
			return
		}
	}
}

func isWhitespace(r rune) bool {
	return unicode.IsSpace(r) || r == ','
} // rune is just a Go "character", works for utf8

func (l *Lexer) NextToken() (Token, error) {
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return Token{Type: TokenEOF}, nil
			}
			return Token{}, err
		}
		if isWhitespace(r) {
			continue
		}

		switch r {
		case '(':
			return Token{Type: TokenLParen, Value: "("}, nil
		case ')':
			return Token{Type: TokenRParen, Value: ")"}, nil
		case '\'':
			return Token{Type: TokenQuote, Value: "'"}, nil
		default:
			// Accumulate symbol
			var sb strings.Builder
			sb.WriteRune(r)
			for {
				nr, _, err := l.reader.ReadRune()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return Token{}, err
				}
				if isWhitespace(nr) || nr == '(' || nr == ')' || nr == '\'' {
					_ = l.reader.UnreadRune()
					break
				}
				sb.WriteRune(nr)
			}
			return Token{Type: TokenSymbol, Value: sb.String()}, nil
		}
	}
}

type Parser struct {
	lexer     *Lexer // pointer to the tokenizer (aka lexer) to get the next token in stream on demand
	peekToken Token  // 1-slot buffer holding the token that has been inspected ahead of time
	hasPeek   bool   // boolean flag whether peekToken currently contains an unconsumed token (true) or if the buffer is empty (false)
}

func NewParser(lexer *Lexer) *Parser {
	return &Parser{lexer: lexer}
}

func (p *Parser) DiscardLine() {
	p.hasPeek = false
	p.lexer.DiscardLine()
}

func (p *Parser) peek() (Token, error) {
	if !p.hasPeek {
		tok, err := p.lexer.NextToken()
		if err != nil {
			return Token{}, err
		}
		p.peekToken = tok
		p.hasPeek = true
	}
	return p.peekToken, nil
}

func (p *Parser) next() (Token, error) {
	tok, err := p.peek()
	if err != nil {
		return Token{}, err
	}
	p.hasPeek = false
	return tok, nil
}

func (p *Parser) ParseExpr() (SExpr, error) {
	tok, err := p.next()
	if err != nil {
		return nil, err
	}
	switch tok.Type {
	case TokenEOF:
		return nil, io.EOF
	case TokenSymbol:
		return Atom{Value: tok.Value}, nil
	case TokenLParen:
		return p.ParseList()
	case TokenRParen:
		return nil, fmt.Errorf("syntax error: unexpected ')'")
	case TokenQuote:
		target, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}
		return Pair{
			Car: Atom{Value: "quote"},
			Cdr: Pair{Car: target, Cdr: nil},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected token: %v", tok)
	}
}

func (p *Parser) ParseList() (SExpr, error) {
	peekTok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if peekTok.Type == TokenRParen {
		_, _ = p.next()
		return nil, nil
	}
	if peekTok.Type == TokenEOF {
		return nil, fmt.Errorf("syntax error: unexpected EOF, expected ')'")
	}

	car, err := p.ParseExpr()
	if err != nil {
		return nil, err
	}
	cdr, err := p.ParseList()
	if err != nil {
		return nil, err
	}
	return Pair{Car: car, Cdr: cdr}, nil
}

// Printer
func FormatSExpr(expr SExpr) string {
	if expr == nil {
		return "()"
	}
	switch v := expr.(type) {
	case Atom:
		return v.Value
	case Pair:
		var sb strings.Builder
		sb.WriteString("(")
		curr := expr
		first := true
		for curr != nil {
			if !first {
				sb.WriteString(" ")
			}
			first = false
			switch node := curr.(type) {
			case Pair:
				sb.WriteString(FormatSExpr(node.Car))
				curr = node.Cdr
			case Atom:
				sb.WriteString(". ")
				sb.WriteString(node.Value)
				curr = nil
			default:
				curr = nil
			}
		}
		sb.WriteString(")")
		return sb.String()
	default:
		return fmt.Sprintf("%v", expr)
	}
}
