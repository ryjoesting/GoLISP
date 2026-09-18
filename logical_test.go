package main

import (
	"strings"
	"testing"
)

func logicalEvalError(t *testing.T, input string) error {
	t.Helper()
	parser := NewParser(NewLexer(strings.NewReader(input)))
	expr, err := parser.ParseExpr()
	if err != nil {
		t.Fatalf("ParseExpr(%q) returned error: %v", input, err)
	}
	_, err = Eval(expr)
	return err
}

func TestAndOrEvaluation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "and returns second value", input: "(and? 'T 'value)", want: "value"},
		{name: "and short circuits false", input: "(and? () 'value)", want: "()"},
		{name: "or returns first value", input: "(or? 'value (unknown x))", want: "value"},
		{name: "or evaluates second after false", input: "(or? () 'value)", want: "value"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := FormatSExpr(evalInput(t, test.input)); got != test.want {
				t.Errorf("Eval(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestAndOrShortCircuitSideEffects(t *testing.T) {
	resetRho(t)

	if got := FormatSExpr(evalAssignmentInput(t, "(and? () (set touched 'yes))")); got != "()" {
		t.Fatalf("and? result = %q, want %q", got, "()")
	}
	if got := FormatSExpr(lookup(Atom{Value: "touched"})); got != "touched" {
		t.Errorf("and? evaluated skipped operand, lookup = %q", got)
	}

	if got := FormatSExpr(evalAssignmentInput(t, "(or? 'T (set touched 'yes))")); got != "T" {
		t.Fatalf("or? result = %q, want %q", got, "T")
	}
	if got := FormatSExpr(lookup(Atom{Value: "touched"})); got != "touched" {
		t.Errorf("or? evaluated skipped operand, lookup = %q", got)
	}
}

func TestAndOrArityErrors(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "(and? 'T)", want: "and? expects 2 argument(s), got 1"},
		{input: "(and? 'T 'T 'T)", want: "and? expects 2 argument(s), got 3"},
		{input: "(or? 'T)", want: "or? expects 2 argument(s), got 1"},
		{input: "(or? 'T 'T 'T)", want: "or? expects 2 argument(s), got 3"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if err := logicalEvalError(t, test.input); err == nil || err.Error() != test.want {
				t.Errorf("Eval(%q) error = %v, want %q", test.input, err, test.want)
			}
		})
	}
}

func TestEq(t *testing.T) {
	resetRho(t)

	if got := FormatSExpr(evalAssignmentInput(t, "(set same 'value)")); got != "()" {
		t.Fatalf("set result = %q, want %q", got, "()")
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "matching atoms", input: "(eq? 'a 'a)", want: "T"},
		{name: "different atoms", input: "(eq? 'a 'b)", want: "()"},
		{name: "evaluated atom", input: "(eq? same 'value)", want: "T"},
		{name: "nil values", input: "(eq? () ())", want: "()"},
		{name: "equal lists", input: "(eq? '(a) '(a))", want: "()"},
		{name: "atom and list", input: "(eq? 'a '(a))", want: "()"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := FormatSExpr(evalAssignmentInput(t, test.input)); got != test.want {
				t.Errorf("Eval(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestEqArityErrors(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "(eq? 'a)", want: "eq? expects 2 argument(s), got 1"},
		{input: "(eq? 'a 'b 'c)", want: "eq? expects 2 argument(s), got 3"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if err := logicalEvalError(t, test.input); err == nil || err.Error() != test.want {
				t.Errorf("Eval(%q) error = %v, want %q", test.input, err, test.want)
			}
		})
	}
}

func TestIfEvaluation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "true branch", input: "(if 'T 'yes 'no)", want: "yes"},
		{name: "false branch", input: "(if () 'yes 'no)", want: "no"},
		{name: "skips false branch", input: "(if 'T 'yes (unknown x))", want: "yes"},
		{name: "skips true branch", input: "(if () (unknown x) 'no)", want: "no"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := FormatSExpr(evalInput(t, test.input)); got != test.want {
				t.Errorf("Eval(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestIfSkipsBranchSideEffects(t *testing.T) {
	resetRho(t)

	evalAssignmentInput(t, "(if 'T 'done (set skipped 'yes))")
	if got := FormatSExpr(lookup(Atom{Value: "skipped"})); got != "skipped" {
		t.Errorf("if evaluated skipped false branch, lookup = %q", got)
	}

	evalAssignmentInput(t, "(if () (set skipped 'yes) 'done)")
	if got := FormatSExpr(lookup(Atom{Value: "skipped"})); got != "skipped" {
		t.Errorf("if evaluated skipped true branch, lookup = %q", got)
	}
}

func TestIfArityErrors(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "(if 'T 'yes)", want: "if expects 3 argument(s), got 2"},
		{input: "(if 'T 'yes 'no 'extra)", want: "if expects 3 argument(s), got 4"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if err := logicalEvalError(t, test.input); err == nil || err.Error() != test.want {
				t.Errorf("Eval(%q) error = %v, want %q", test.input, err, test.want)
			}
		})
	}
}

func TestCondEvaluation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "first match", input: "(cond ('T 'first 'T 'second))", want: "first"},
		{name: "later match", input: "(cond (() 'first 'T 'second))", want: "second"},
		{name: "no match", input: "(cond (() 'first () 'second))", want: "()"},
		{name: "evaluates selected result", input: "(cond (() 'first 'T (cons 'a '())))", want: "(a)"},
		{name: "skips later error", input: "(cond ('T 'first (unknown x) 'second))", want: "first"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := FormatSExpr(evalInput(t, test.input)); got != test.want {
				t.Errorf("Eval(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestCondShortCircuitsSideEffects(t *testing.T) {
	resetRho(t)

	if got := FormatSExpr(evalAssignmentInput(t, "(cond ((set checked 'yes) 'first 'T 'second))")); got != "second" {
		t.Fatalf("cond result = %q, want %q", got, "second")
	}
	if got := FormatSExpr(lookup(Atom{Value: "checked"})); got != "yes" {
		t.Errorf("cond did not evaluate predicate, lookup = %q", got)
	}

	if got := FormatSExpr(evalAssignmentInput(t, "(cond ('T 'first (set skipped 'yes) 'second))")); got != "first" {
		t.Fatalf("cond result = %q, want %q", got, "first")
	}
	if got := FormatSExpr(lookup(Atom{Value: "skipped"})); got != "skipped" {
		t.Errorf("cond evaluated skipped clause, lookup = %q", got)
	}
}

func TestCondErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "missing argument", input: "(cond)", want: "cond expects 1 argument(s), got 0"},
		{name: "extra argument", input: "(cond () ())", want: "cond expects 1 argument(s), got 2"},
		{name: "odd clause count", input: "(cond ('T))", want: "cond expects predicate/result pairs, got 1 expression(s)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := logicalEvalError(t, test.input); err == nil || err.Error() != test.want {
				t.Errorf("Eval(%q) error = %v, want %q", test.input, err, test.want)
			}
		})
	}
}

func TestCondRejectsImproperClauses(t *testing.T) {
	args := []SExpr{
		Pair{
			Car: Atom{Value: "T"},
			Cdr: Atom{Value: "tail"},
		},
	}

	_, err := evalCond(args)
	if err == nil || err.Error() != "cond clauses must form a list: function arguments must form a list" {
		t.Errorf("evalCond() error = %v, want improper-clause error", err)
	}
}
