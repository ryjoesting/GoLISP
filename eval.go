package main

import "fmt"

func car(expr SExpr) (SExpr, error) {
	pair, ok := expr.(Pair)
	if !ok {
		return nil, fmt.Errorf("car expects a list")
	}
	return pair.Car, nil
}

func cdr(expr SExpr) (SExpr, error) {
	pair, ok := expr.(Pair)
	if !ok {
		return nil, fmt.Errorf("cdr expects a list")
	}
	return pair.Cdr, nil
}

func cons(first, rest SExpr) SExpr {
	return Pair{Car: first, Cdr: rest}
}

func quote(expr SExpr) SExpr {
	return expr
}

// collectArgs reads the proper list of arguments from a call.
func collectArgs(expr SExpr) ([]SExpr, error) {
	var args []SExpr
	for expr != nil {
		pair, ok := expr.(Pair)
		if !ok {
			return nil, fmt.Errorf("function arguments must form a list")
		}
		args = append(args, pair.Car)
		expr = pair.Cdr
	}
	return args, nil
}

func requireArgs(name string, args []SExpr, count int) error {
	if len(args) != count {
		return fmt.Errorf("%s expects %d argument(s), got %d", name, count, len(args))
	}
	return nil
}

func evaluateArgs(args []SExpr) ([]SExpr, error) {
	values := make([]SExpr, len(args))
	for index, arg := range args {
		value, err := Eval(arg)
		if err != nil {
			return nil, err
		}
		values[index] = value
	}
	return values, nil
}

func Eval(expr SExpr) (SExpr, error) {
	if expr == nil {
		return nil, nil
	}
	if atom, ok := expr.(Atom); ok {
		return lookup(atom), nil
	}

	call, ok := expr.(Pair)
	if !ok {
		return expr, nil
	}

	name, ok := call.Car.(Atom)
	if !ok {
		return nil, fmt.Errorf("function name must be an atom")
	}

	args, err := collectArgs(call.Cdr)
	if err != nil {
		return nil, err
	}

	// quote keeps its argument as data.
	if name.Value == "quote" {
		if err := requireArgs(name.Value, args, 1); err != nil {
			return nil, err
		}
		return quote(args[0]), nil
	}
	if name.Value == "eval" {
		if err := requireArgs(name.Value, args, 1); err != nil {
			return nil, err
		}
		return Eval(args[0])
	}
	if name.Value == "set" {
		if err := requireArgs(name.Value, args, 2); err != nil {
			return nil, err
		}
		variable, ok := args[0].(Atom)
		if !ok {
			return nil, fmt.Errorf("set expects an atom name")
		}
		value, err := Eval(args[1])
		if err != nil {
			return nil, err
		}
		assign(variable, value)
		return nil, nil
	}
	if value, handled, err := evalLogical(name.Value, args); handled {
		return value, err
	}
	if value, handled, err := evalMath(name.Value, args); handled {
		return value, err
	}

	values, err := evaluateArgs(args)
	if err != nil {
		return nil, err
	}

	switch name.Value {
	case "car":
		if err := requireArgs(name.Value, values, 1); err != nil {
			return nil, err
		}
		return car(values[0])
	case "cdr":
		if err := requireArgs(name.Value, values, 1); err != nil {
			return nil, err
		}
		return cdr(values[0])
	case "cons":
		if err := requireArgs(name.Value, values, 2); err != nil {
			return nil, err
		}
		return cons(values[0], values[1]), nil
	case "nil?":
		if err := requireArgs(name.Value, values, 1); err != nil {
			return nil, err
		}
		return isNil(values[0]), nil
	case "atom?":
		if err := requireArgs(name.Value, values, 1); err != nil {
			return nil, err
		}
		return isAtom(values[0]), nil
	case "list?":
		if err := requireArgs(name.Value, values, 1); err != nil {
			return nil, err
		}
		return isList(values[0]), nil
	default:
		return nil, fmt.Errorf("unknown function: %s", name.Value)
	}
}
