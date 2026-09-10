package main

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestParseAndFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		// Assignment-specified cases
		{name: "empty list", input: "()", expected: "()", wantErr: false},
		{name: "simple symbol", input: "a", expected: "a", wantErr: false},
		{name: "number symbol", input: "45", expected: "45", wantErr: false},
		{name: "basic list", input: "(1 2 3)", expected: "(1 2 3)", wantErr: false},
		{name: "nested list", input: "((1 2 3) (a b c))", expected: "((1 2 3) (a b c))", wantErr: false},
		{name: "list of empty lists", input: "( ( ) ( ) ( ) )", expected: "(() () ())", wantErr: false},
		{
			name:     "complex assignment example",
			input:    "(() a 45 test (1 2 3) ((1 2 3) (a b c) ) ( ( ) ( ) ( ) ))",
			expected: "(() a 45 test (1 2 3) ((1 2 3) (a b c)) (() () ()))",
			wantErr:  false,
		},

		// Lexer nuance cases
		{name: "commas as whitespace", input: "(a, b, c)", expected: "(a b c)", wantErr: false},
		{name: "multiline expression", input: "(\n  a\n  b\n  c\n)", expected: "(a b c)", wantErr: false},

		// Error cases
		{name: "unmatched open paren", input: "(a b", wantErr: true},
		{name: "unexpected close paren", input: ")", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))
			parser := NewParser(lexer)

			expr, err := parser.ParseExpr()
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseExpr() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			got := FormatSExpr(expr)
			if got != tt.expected {
				t.Errorf("FormatSExpr() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestInternalStructure(t *testing.T) {
	input := "(a b)"
	p := NewParser(NewLexer(strings.NewReader(input)))
	expr, err := p.ParseExpr()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Root should be a Pair
	p1, ok := expr.(Pair)
	if !ok {
		t.Fatalf("expected Pair, got %T", expr)
	}

	// Car should be Atom("a")
	car1, ok := p1.Car.(Atom)
	if !ok || car1.Value != "a" {
		t.Errorf("expected Car to be Atom('a'), got %v", p1.Car)
	}

	// Cdr should be another Pair
	p2, ok := p1.Cdr.(Pair)
	if !ok {
		t.Fatalf("expected Cdr to be Pair, got %T", p1.Cdr)
	}

	// Tail of list should be nil
	if p2.Cdr != nil {
		t.Errorf("expected end of list to be nil, got %v", p2.Cdr)
	}
}

func TestMultipleSExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "multiple on one line (assignment example)",
			input:    "() a 45 test",
			expected: []string{"()", "a", "45", "test"},
		},
		{
			name: "multiple spanning multiple lines",
			input: `
				(1 2 3)
				((1 2 3) (a b c))
				( ( ) ( ) ( ) )
			`,
			expected: []string{
				"(1 2 3)",
				"((1 2 3) (a b c))",
				"(() () ())",
			},
		},
		{
			name:     "symbols with operators and nil",
			input:    "(+ - nil * /)",
			expected: []string{"(+ - nil * /)"},
		},
		{
			name:     "commas between multiple top-level expressions",
			input:    "a, b, (c d), e",
			expected: []string{"a", "b", "(c d)", "e"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))
			parser := NewParser(lexer)

			var results []string
			for {
				expr, err := parser.ParseExpr()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					t.Fatalf("unexpected error during parse: %v", err)
				}
				results = append(results, FormatSExpr(expr))
			}

			if len(results) != len(tt.expected) {
				t.Fatalf("got %d expressions, want %d (got: %v)", len(results), len(tt.expected), results)
			}

			for i := range results {
				if results[i] != tt.expected[i] {
					t.Errorf("expr[%d] = %q, want %q", i, results[i], tt.expected[i])
				}
			}
		})
	}
}
