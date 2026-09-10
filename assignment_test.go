package main

import (
	"strings"
	"testing"
)

func resetRho(t *testing.T) {
	t.Helper()
	previous := rho
	rho = nil
	t.Cleanup(func() {
		rho = previous
	})
}

func evalAssignmentInput(t *testing.T, input string) SExpr {
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

func TestAssignmentTruthValues(t *testing.T) {
	if got := FormatSExpr(truthValue(true)); got != "T" {
		t.Errorf("truthValue(true) = %q, want %q", got, "T")
	}
	if got := FormatSExpr(truthValue(false)); got != "()" {
		t.Errorf("truthValue(false) = %q, want %q", got, "()")
	}

	atom := Atom{Value: "a"}
	list := Pair{Car: atom, Cdr: nil}

	for _, test := range []struct {
		name string
		got  SExpr
		want string
	}{
		{name: "nil? empty", got: isNil(nil), want: "T"},
		{name: "nil? atom", got: isNil(atom), want: "()"},
		{name: "nil? list", got: isNil(list), want: "()"},
		{name: "atom? atom", got: isAtom(atom), want: "T"},
		{name: "atom? list", got: isAtom(list), want: "()"},
		{name: "atom? empty undefined case", got: isAtom(nil), want: "()"},
		{name: "list? atom", got: isList(atom), want: "()"},
		{name: "list? list", got: isList(list), want: "T"},
		{name: "list? empty undefined case", got: isList(nil), want: "()"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := FormatSExpr(test.got); got != test.want {
				t.Errorf("%s = %q, want %q", test.name, got, test.want)
			}
		})
	}
}

func TestAssignBuildsProperEnvironment(t *testing.T) {
	resetRho(t)

	assign(Atom{Value: "a"}, Atom{Value: "2"})
	if got := FormatSExpr(rho); got != "((a 2))" {
		t.Fatalf("rho after one assignment = %q, want %q", got, "((a 2))")
	}

	environment, ok := rho.(Pair)
	if !ok {
		t.Fatalf("rho has type %T, want Pair", rho)
	}
	binding, ok := environment.Car.(Pair)
	if !ok {
		t.Fatalf("binding has type %T, want Pair", environment.Car)
	}
	if binding.Cdr == nil {
		t.Fatal("binding has no value cell")
	}
	valueCell, ok := binding.Cdr.(Pair)
	if !ok || valueCell.Cdr != nil {
		t.Fatalf("binding value is not a proper two-element list: %s", FormatSExpr(binding))
	}
	if environment.Cdr != nil {
		t.Fatalf("rho tail = %s, want ()", FormatSExpr(environment.Cdr))
	}

	assign(Atom{Value: "a"}, Atom{Value: "4"})
	assign(Atom{Value: "b"}, Atom{Value: "6"})
	if got := FormatSExpr(rho); got != "((b 6) (a 4) (a 2))" {
		t.Errorf("rho after repeated assignments = %q, want %q", got, "((b 6) (a 4) (a 2))")
	}
}

func TestLookup(t *testing.T) {
	resetRho(t)

	name := Atom{Value: "a"}
	if got := FormatSExpr(lookup(name)); got != "a" {
		t.Errorf("lookup of unbound atom = %q, want %q", got, "a")
	}

	assign(name, Atom{Value: "2"})
	assign(Atom{Value: "items"}, Pair{Car: Atom{Value: "x"}, Cdr: nil})
	if got := FormatSExpr(lookup(name)); got != "2" {
		t.Errorf("lookup of bound atom = %q, want %q", got, "2")
	}
	if got := FormatSExpr(lookup(Atom{Value: "items"})); got != "(x)" {
		t.Errorf("lookup of list binding = %q, want %q", got, "(x)")
	}

	assign(name, Atom{Value: "4"})
	if got := FormatSExpr(lookup(name)); got != "4" {
		t.Errorf("lookup after shadowing = %q, want %q", got, "4")
	}
	if got := FormatSExpr(rho); got != "((a 4) (items (x)) (a 2))" {
		t.Errorf("rho after shadowing = %q, want %q", got, "((a 4) (items (x)) (a 2))")
	}
}

func TestAssignmentEvaluation(t *testing.T) {
	resetRho(t)

	if got := FormatSExpr(evalAssignmentInput(t, "(set a 2)")); got != "()" {
		t.Errorf("set result = %q, want %q", got, "()")
	}
	if got := FormatSExpr(evalAssignmentInput(t, "a")); got != "2" {
		t.Errorf("lookup after set = %q, want %q", got, "2")
	}

	evalAssignmentInput(t, "(set list (cons (quote x) (quote ())))")
	if got := FormatSExpr(evalAssignmentInput(t, "list")); got != "(x)" {
		t.Errorf("evaluated assignment value = %q, want %q", got, "(x)")
	}

	evalAssignmentInput(t, "(set quoted '(x y))")
	if got := FormatSExpr(evalAssignmentInput(t, "quoted")); got != "(x y)" {
		t.Errorf("quoted assignment value = %q, want %q", got, "(x y)")
	}
	if got := FormatSExpr(evalAssignmentInput(t, "unknown")); got != "unknown" {
		t.Errorf("unbound symbol = %q, want %q", got, "unknown")
	}
}

func TestAssignmentPredicates(t *testing.T) {
	resetRho(t)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "nil true", input: "(nil? ())", want: "T"},
		{name: "nil false", input: "(nil? 'T)", want: "()"},
		{name: "atom true", input: "(atom? x)", want: "T"},
		{name: "atom false", input: "(atom? (quote (x)))", want: "()"},
		{name: "list atom false", input: "(list? 'T)", want: "()"},
		{name: "list unbound atom false", input: "(list? x)", want: "()"},
		{name: "list true", input: "(list? (quote (x)))", want: "T"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := FormatSExpr(evalAssignmentInput(t, test.input)); got != test.want {
				t.Errorf("Eval(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestAssignmentErrors(t *testing.T) {
	resetRho(t)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "set missing value", input: "(set a)", want: "set expects 2 argument(s), got 1"},
		{name: "set extra argument", input: "(set a 2 3)", want: "set expects 2 argument(s), got 3"},
		{name: "set non-atom name", input: "(set (a) 2)", want: "set expects an atom name"},
		{name: "set value error", input: "(set a (unknown 2))", want: "unknown function: unknown"},
		{name: "nil predicate arity", input: "(nil?)", want: "nil? expects 1 argument(s), got 0"},
		{name: "atom predicate arity", input: "(atom? a b)", want: "atom? expects 1 argument(s), got 2"},
		{name: "list predicate arity", input: "(list? a b)", want: "list? expects 1 argument(s), got 2"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
