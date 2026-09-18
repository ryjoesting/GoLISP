package main

var rho SExpr
var localEnvironment SExpr

func lookup(name Atom) SExpr {
	if value, found := lookupLocal(name); found {
		return value
	}

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

func lookupLocal(name Atom) (SExpr, bool) {
	for stack := localEnvironment; stack != nil; {
		entry, ok := stack.(Pair)
		if !ok {
			break
		}

		frame, ok := entry.Car.(Pair)
		if ok {
			if value, found := lookupFrame(name, frame); found {
				return value, true
			}
		}
		stack = entry.Cdr
	}
	return nil, false
}

func lookupFrame(name Atom, frame Pair) (SExpr, bool) {
	names := frame.Car
	valueCell, ok := frame.Cdr.(Pair)
	if !ok || valueCell.Cdr != nil {
		return nil, false
	}
	values := valueCell.Car

	for names != nil && values != nil {
		nameCell, ok := names.(Pair)
		if !ok {
			return nil, false
		}
		valueCell, ok := values.(Pair)
		if !ok {
			return nil, false
		}

		localName, ok := nameCell.Car.(Atom)
		if !ok {
			return nil, false
		}
		if localName.Value == name.Value {
			return valueCell.Car, true
		}
		names = nameCell.Cdr
		values = valueCell.Cdr
	}

	return nil, false
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
