package main

import (
	"strings"
	"testing"
)

func TestMathEvaluation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "add", input: "(add 2 3)", want: "5"},
		{name: "sub", input: "(sub 2 3)", want: "-1"},
		{name: "mul", input: "(mul 2 3)", want: "6"},
		{name: "div integer division", input: "(div 7 2)", want: "3"},
		{name: "rem", input: "(rem 7 2)", want: "1"},
		{name: "lt true", input: "(lt 2 3)", want: "T"},
		{name: "lt false", input: "(lt 3 2)", want: "()"},
		{name: "lt equal", input: "(lt 3 3)", want: "()"},
		{name: "nested", input: "(add (mul 2 3) 4)", want: "10"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := FormatSExpr(evalInput(t, test.input)); got != test.want {
				t.Errorf("Eval(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestMathErrors(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "(add 1)", want: "add expects 2 argument(s), got 1"},
		{input: "(sub 1 2 3)", want: "sub expects 2 argument(s), got 3"},
		{input: "(mul a 2)", want: "math functions expect numbers"},
		{input: "(div 4 0)", want: "div cannot divide by zero"},
		{input: "(rem 4 0)", want: "rem cannot divide by zero"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			parser := NewParser(NewLexer(strings.NewReader(test.input)))
			expr, err := parser.ParseExpr()
			if err != nil {
				t.Fatalf("ParseExpr(%q) returned error: %v", test.input, err)
			}
			_, err = Eval(expr)
			if err == nil || err.Error() != test.want {
				t.Errorf("Eval(%q) error = %v, want %q", test.input, err, test.want)
			}
		})
	}
}
