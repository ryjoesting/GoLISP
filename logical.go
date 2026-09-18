package main

import "fmt"

func evalAnd(args []SExpr) (SExpr, error) {
	if err := requireArgs("and?", args, 2); err != nil {
		return nil, err
	}

	first, err := Eval(args[0])
	if err != nil {
		return nil, err
	}
	if first == nil {
		return nil, nil
	}
	return Eval(args[1])
}

func evalOr(args []SExpr) (SExpr, error) {
	if err := requireArgs("or?", args, 2); err != nil {
		return nil, err
	}

	first, err := Eval(args[0])
	if err != nil {
		return nil, err
	}
	if first != nil {
		return first, nil
	}
	return Eval(args[1])
}

func evalIf(args []SExpr) (SExpr, error) {
	if err := requireArgs("if", args, 3); err != nil {
		return nil, err
	}

	condition, err := Eval(args[0])
	if err != nil {
		return nil, err
	}
	if condition == nil {
		return Eval(args[2])
	}
	return Eval(args[1])
}

func evalCond(args []SExpr) (SExpr, error) {
	if err := requireArgs("cond", args, 1); err != nil {
		return nil, err
	}

	clauses, err := collectArgs(args[0])
	if err != nil {
		return nil, fmt.Errorf("cond clauses must form a list: %w", err)
	}
	if len(clauses)%2 != 0 {
		return nil, fmt.Errorf("cond expects predicate/result pairs, got %d expression(s)", len(clauses))
	}

	for index := 0; index < len(clauses); index += 2 {
		condition, err := Eval(clauses[index])
		if err != nil {
			return nil, err
		}
		if condition != nil {
			return Eval(clauses[index+1])
		}
	}
	return nil, nil
}

func evalLogical(name string, args []SExpr) (SExpr, bool, error) {
	switch name {
	case "and?":
		value, err := evalAnd(args)
		return value, true, err
	case "or?":
		value, err := evalOr(args)
		return value, true, err
	case "if":
		value, err := evalIf(args)
		return value, true, err
	case "cond":
		value, err := evalCond(args)
		return value, true, err
	case "eq?":
		values, err := evaluateArgs(args)
		if err != nil {
			return nil, true, err
		}
		if err := requireArgs(name, values, 2); err != nil {
			return nil, true, err
		}
		left, leftIsAtom := values[0].(Atom)
		right, rightIsAtom := values[1].(Atom)
		return truthValue(leftIsAtom && rightIsAtom && left.Value == right.Value), true, nil
	default:
		return nil, false, nil
	}
}
