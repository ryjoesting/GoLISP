package main

import (
	"fmt"
	"strconv"
)

func numberValue(value SExpr) (int, error) {
	atom, ok := value.(Atom)
	if !ok {
		return 0, fmt.Errorf("math functions expect numbers")
	}

	number, err := strconv.Atoi(atom.Value)
	if err != nil {
		return 0, fmt.Errorf("math functions expect numbers")
	}
	return number, nil
}

func mathArgs(name string, args []SExpr) ([2]int, error) {
	if err := requireArgs(name, args, 2); err != nil {
		return [2]int{}, err
	}

	values := [2]int{}
	for index, arg := range args {
		value, err := numberValue(arg)
		if err != nil {
			return [2]int{}, err
		}
		values[index] = value
	}
	return values, nil
}

func mathResult(name string, args []SExpr, operation func(int, int) (int, error)) (SExpr, error) {
	values, err := mathArgs(name, args)
	if err != nil {
		return nil, err
	}

	result, err := operation(values[0], values[1])
	if err != nil {
		return nil, err
	}
	return Atom{Value: strconv.Itoa(result)}, nil
}

func add(args []SExpr) (SExpr, error) {
	return mathResult("add", args, func(left, right int) (int, error) {
		return left + right, nil
	})
}

func sub(args []SExpr) (SExpr, error) {
	return mathResult("sub", args, func(left, right int) (int, error) {
		return left - right, nil
	})
}

func mul(args []SExpr) (SExpr, error) {
	return mathResult("mul", args, func(left, right int) (int, error) {
		return left * right, nil
	})
}

func div(args []SExpr) (SExpr, error) {
	return mathResult("div", args, func(left, right int) (int, error) {
		if right == 0 {
			return 0, fmt.Errorf("div cannot divide by zero")
		}
		return left / right, nil
	})
}

func rem(args []SExpr) (SExpr, error) {
	return mathResult("rem", args, func(left, right int) (int, error) {
		if right == 0 {
			return 0, fmt.Errorf("rem cannot divide by zero")
		}
		return left % right, nil
	})
}

func lt(args []SExpr) (SExpr, error) {
	values, err := mathArgs("lt", args)
	if err != nil {
		return nil, err
	}
	return truthValue(values[0] < values[1]), nil
}

func evalMath(name string, args []SExpr) (SExpr, bool, error) {
	switch name {
	case "add", "sub", "mul", "div", "rem", "lt":
	default:
		return nil, false, nil
	}

	values, err := evaluateArgs(args)
	if err != nil {
		return nil, true, err
	}

	var value SExpr
	switch name {
	case "add":
		value, err = add(values)
	case "sub":
		value, err = sub(values)
	case "mul":
		value, err = mul(values)
	case "div":
		value, err = div(values)
	case "rem":
		value, err = rem(values)
	case "lt":
		value, err = lt(values)
	}
	return value, true, err
}
