package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func evalInput(t *testing.T, input string) SExpr {
	t.Helper()

	parser := NewParser(NewLexer(strings.NewReader(input)))
	expr, err := parser.ParseExpr()
	if err != nil {
		t.Fatalf("ParseExpr(%q) returned error: %v", input, err)
	}

	result, err := Eval(expr)
	if err != nil {
		t.Fatalf("Eval(%q) returned error: %v", input, err)
	}
	return result
}

func TestCarAndCdr(t *testing.T) {
	list := Pair{
		Car: Atom{Value: "a"},
		Cdr: Pair{Car: Atom{Value: "b"}, Cdr: nil},
	}

	gotCar, err := car(list)
	if err != nil {
		t.Fatalf("car() returned error: %v", err)
	}
	if got := FormatSExpr(gotCar); got != "a" {
		t.Errorf("car() = %q, want %q", got, "a")
	}

	gotCdr, err := cdr(list)
	if err != nil {
		t.Fatalf("cdr() returned error: %v", err)
	}
	if got := FormatSExpr(gotCdr); got != "(b)" {
		t.Errorf("cdr() = %q, want %q", got, "(b)")
	}

	withImproperTail := Pair{Car: Atom{Value: "a"}, Cdr: Atom{Value: "b"}}
	gotCdr, err = cdr(withImproperTail)
	if err != nil {
		t.Fatalf("cdr() with improper tail returned error: %v", err)
	}
	if got := FormatSExpr(gotCdr); got != "b" {
		t.Errorf("cdr() with improper tail = %q, want %q", got, "b")
	}
}

func TestCarAndCdrRejectNonPairs(t *testing.T) {
	tests := []struct {
		name string
		call func(SExpr) (SExpr, error)
		want string
	}{
		{name: "car", call: car, want: "car expects a list"},
		{name: "cdr", call: cdr, want: "cdr expects a list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, input := range []SExpr{Atom{Value: "a"}, nil} {
				_, err := tt.call(input)
				if err == nil || err.Error() != tt.want {
					t.Errorf("%s(%v) error = %v, want %q", tt.name, input, err, tt.want)
				}
			}
		})
	}
}

func TestCons(t *testing.T) {
	proper := cons(Atom{Value: "a"}, nil)
	if got := FormatSExpr(proper); got != "(a)" {
		t.Errorf("cons(a, nil) = %q, want %q", got, "(a)")
	}

	nested := cons(Atom{Value: "a"}, cons(Atom{Value: "b"}, nil))
	if got := FormatSExpr(nested); got != "(a b)" {
		t.Errorf("nested cons() = %q, want %q", got, "(a b)")
	}

	dotted := cons(Atom{Value: "a"}, Atom{Value: "b"})
	if got := FormatSExpr(dotted); got != "(a . b)" {
		t.Errorf("dotted cons() = %q, want %q", got, "(a . b)")
	}
}

func TestQuote(t *testing.T) {
	list := Pair{Car: Atom{Value: "a"}, Cdr: nil}
	for _, input := range []SExpr{Atom{Value: "a"}, list, nil} {
		if got := quote(input); got != input {
			t.Errorf("quote(%v) = %v, want the original value", input, got)
		}
	}
}

func TestEvalExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty list", input: "()", want: "()"},
		{name: "atom", input: "a", want: "a"},
		{name: "car", input: "(car (quote (a b c)))", want: "a"},
		{name: "cdr", input: "(cdr (quote (a b c)))", want: "(b c)"},
		{name: "cons list", input: "(cons a ())", want: "(a)"},
		{name: "cons dotted pair", input: "(cons a b)", want: "(a . b)"},
		{name: "quote list", input: "(quote (car x))", want: "(car x)"},
		{name: "shorthand quote", input: "'(car x)", want: "(car x)"},
		{name: "nested eval car", input: "(eval (car (quote (a b c))))", want: "a"},
		{name: "nested eval cdr", input: "(eval (cdr (quote (a b c))))", want: "(b c)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatSExpr(evalInput(t, tt.input)); got != tt.want {
				t.Errorf("Eval(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEvalErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "unknown function", input: "(unknown a)", want: "unknown function: unknown"},
		{name: "non-atom function head", input: "((a) b)", want: "function name must be an atom"},
		{name: "car missing argument", input: "(car)", want: "car expects 1 argument(s), got 0"},
		{name: "cdr extra argument", input: "(cdr a b)", want: "cdr expects 1 argument(s), got 2"},
		{name: "cons missing argument", input: "(cons a)", want: "cons expects 2 argument(s), got 1"},
		{name: "quote extra argument", input: "(quote a b)", want: "quote expects 1 argument(s), got 2"},
		{name: "eval missing argument", input: "(eval)", want: "eval expects 1 argument(s), got 0"},
		{name: "car atom", input: "(car a)", want: "car expects a list"},
		{name: "cdr empty list", input: "(cdr ())", want: "cdr expects a list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(NewLexer(strings.NewReader(tt.input)))
			expr, err := parser.ParseExpr()
			if err != nil {
				t.Fatalf("ParseExpr(%q) returned error: %v", tt.input, err)
			}
			_, err = Eval(expr)
			if err == nil || err.Error() != tt.want {
				t.Errorf("Eval(%q) error = %v, want %q", tt.input, err, tt.want)
			}
		})
	}

	_, err := Eval(Pair{Car: Atom{Value: "car"}, Cdr: Atom{Value: "a"}})
	if err == nil || err.Error() != "function arguments must form a list" {
		t.Errorf("Eval() improper arguments error = %v, want %q", err, "function arguments must form a list")
	}
}

func TestQuoteDoesNotEvaluateUnknownCall(t *testing.T) {
	result := evalInput(t, "'(unknown a)")
	if got := FormatSExpr(result); got != "(unknown a)" {
		t.Errorf("quoted unknown call = %q, want %q", got, "(unknown a)")
	}
}

func TestREPLIntegration(t *testing.T) {
	command := exec.Command("go", "run", ".")
	command.Stdin = strings.NewReader("(car '(a b))\n(cons a b)\n(unknown a) trailing\na\n")

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("REPL process returned an unexpected error: %v", err)
	}

	if got, want := stdout.String(), "a\n(a . b)\na\n"; got != want {
		t.Errorf("REPL stdout = %q, want %q", got, want)
	}
	if !strings.Contains(stderr.String(), "Error: unknown function: unknown") {
		t.Errorf("REPL stderr = %q, want an unknown-function error", stderr.String())
	}
	if !strings.HasSuffix(stdout.String(), "a\n") {
		t.Errorf("REPL did not continue with the next line: %q", stdout.String())
	}
}
