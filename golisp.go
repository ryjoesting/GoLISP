package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func main() {
	parser := NewParser(NewLexer(os.Stdin))

	for {
		expr, err := parser.ParseExpr()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}

		result, err := Eval(expr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			parser.DiscardLine()
			continue
		}
		fmt.Println(FormatSExpr(result))
	}
}
