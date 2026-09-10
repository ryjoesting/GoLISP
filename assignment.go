package main

var rho SExpr

func lookup(name Atom) SExpr {
	for environment := rho; environment != nil; {
		entry, ok := environment.(Pair)
		if !ok {
			return name
		}

		binding, ok := entry.Car.(Pair)
		if !ok {
			environment = entry.Cdr
			continue
		}

		bindingName, ok := binding.Car.(Atom)
		if !ok {
			environment = entry.Cdr
			continue
		}

		value, ok := binding.Cdr.(Pair)
		if ok && bindingName.Value == name.Value {
			return value.Car
		}
		environment = entry.Cdr
	}
	return name
}

func assign(name Atom, value SExpr) {
	binding := Pair{
		Car: name,
		Cdr: Pair{Car: value, Cdr: nil},
	}
	rho = Pair{Car: binding, Cdr: rho}
}

func truthValue(value bool) SExpr {
	if value {
		return Atom{Value: "T"}
	}
	return nil
}

func isNil(value SExpr) SExpr {
	return truthValue(value == nil)
}

func isAtom(value SExpr) SExpr {
	if value == nil {
		// The assignment leaves atom? on () undefined.
		return nil
	}
	_, ok := value.(Atom)
	return truthValue(ok)
}

func isList(value SExpr) SExpr {
	if value == nil {
		// The assignment leaves list? on () undefined.
		return nil
	}
	_, ok := value.(Pair)
	return truthValue(ok)
}
