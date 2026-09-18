# GoLISP

GoLISP is a small Lisp interpreter written in Go. It reads S-expressions from standard input, evaluates them, and prints one result for each expression.

The project currently includes:

- An S-expression lexer, parser, and formatter
- Lists, atoms, empty lists, and quote shorthand
- Core list operations
- Explicit evaluation with `eval`
- Global assignment and symbol lookup
- Basic predicates
- A line-oriented REPL with evaluation-error recovery

## Requirements

- Go 1.27.1 or a compatible newer Go version
- A terminal or shell

The project uses only the Go standard library and has no external dependencies.

## Clone and Run

Clone the repository and move into the project directory:

```text
git clone <repository-url>
cd golisp
```

Run the interpreter with:

```text
go run .
```

GoLISP reads from standard input until it reaches end-of-file. There is no interactive prompt, so type or pipe expressions into the process. On Windows, press `Ctrl+Z` followed by Enter to send end-of-file. On macOS or Linux, press `Ctrl+D`.

Example using a file or a shell pipeline:

```text
echo "(car '(a b c))" | go run .
```

Output:

```text
a
```

You can also build an executable:

```text
go build .
```

Then run the generated executable from the project directory.

## Running Tests

Run all tests with:

```text
go test ./...
```

The tests cover parsing, formatting, evaluation, assignment, predicates, error handling, and REPL behavior.

## Language Basics

### Atoms

An atom is a symbol represented by text:

```text
a
hello
45
```

Number-looking values such as `45` are currently ordinary symbols. The interpreter does not perform numeric conversion or arithmetic. The symbol `nil` is also an ordinary symbol; the empty list is written as `()`.

### Lists

Lists are written with parentheses:

```text
(a b c)
((a b) (c d))
()
```

Whitespace separates elements. Commas are also treated as whitespace, so `(a, b, c)` is accepted as `(a b c)`. Multiple expressions may be placed on one line or across multiple lines.

The reader supports quote shorthand:

```text
'a
'(a b c)
```

These are parsed as:

```text
(quote a)
(quote (a b c))
```

A dotted pair can be created with `cons` and printed, but dotted-pair input syntax is not supported by the parser.

## Evaluation Rules

- The empty list `()` evaluates to `()`.
- A standalone atom is looked up in the global environment. An unbound atom evaluates to itself.
- A non-empty list is treated as a function call.
- The first element of a call must be an atom naming a supported built-in function.
- Ordinary function arguments are evaluated before the function runs.
- `quote`, `eval`, and `set` have special evaluation rules.

Function names are not general function values. For example, `a` performs symbol lookup, while `(a)` attempts to call a zero-argument function named `a`. Since only built-in functions are supported, `(a)` produces an unknown-function error unless that behavior is changed in a future version.

## Built-in Functions

### `car`

Returns the first element of a pair:

```text
(car '(a b c))
```

Result:

```text
a
```

### `cdr`

Returns the rest of a pair:

```text
(cdr '(a b c))
```

Result:

```text
(b c)
```

`car` and `cdr` currently report an error when given an atom or the empty list.

### `cons`

Constructs a pair from two expressions:

```text
(cons 'a '())
(cons 'a '(b c))
(cons 'a 'b)
```

Results:

```text
(a)
(a b c)
(a . b)
```

### `quote`

Returns its argument without evaluating it:

```text
(quote (car x))
```

Result:

```text
(car x)
```

The shorter equivalent form is `'(car x)`.

### `eval`

Evaluates its argument as an expression:

```text
(eval (car '(a b c)))
```

Result:

```text
a
```

### `set`

Creates a global binding:

```text
(set answer 42)
answer
```

Results:

```text
()
42
```

The first argument must be an atom name. The value is evaluated before it is stored. Assignment prepends a new binding rather than replacing an old one, so the newest binding shadows earlier bindings:

```text
(set answer 2)
(set answer 4)
answer
```

The final expression evaluates to `4`.

There are no local environments or function parameters yet.

### `nil?`

Returns `T` for the empty list and `()` for other values:

```text
(nil? ())
(nil? 'T)
```

Results:

```text
T
()
```

### `atom?`

Returns `T` for an atom and `()` for a non-empty list:

```text
(atom? x)
(atom? '(x))
```

Results:

```text
T
()
```

The behavior of `(atom? ())` is undefined by the assignment. The current implementation returns `()`.

### `list?`

Returns `T` for a non-empty list and `()` for an atom:

```text
(list? '(x))
(list? x)
```

Results:

```text
T
()
```

The behavior of `(list? ())` is undefined by the assignment. The current implementation returns `()`.

### `and?` and `or?`

These operators evaluate only the expressions needed to determine the result:

```text
(and? 'T 'a)
(or? () 'a)
```

`and?` returns `()` immediately when its first argument is nil; otherwise it returns the evaluated second argument. `or?` returns the evaluated first argument immediately when it is non-nil; otherwise it evaluates and returns the second argument. Both operators require exactly two arguments.

### `eq?`

`eq?` evaluates two arguments and compares them only when both results are atoms. Matching atom values return `T`; nil and pairs return `()` even when two pairs have identical contents.

### `if`

`if` requires a condition and two branches. It evaluates the condition, then evaluates only the true branch when the condition is non-nil or only the false branch when it is nil.

### `cond`

`cond` takes one list containing alternating condition and result expressions:

```text
(cond (() 'first 'T 'fallback))
```

Conditions are evaluated from left to right. The result paired with the first non-nil condition is evaluated and returned. If no condition succeeds, `cond` returns `()`. Its clause list must be proper and contain an even number of expressions.

## REPL Error Behavior

Successful results are printed to standard output. Evaluation errors are printed to standard error with an `Error:` prefix. After an evaluation error, GoLISP discards the remainder of the current input line and continues with the next line.

For example, the input:

```text
(unknown value) ignored-on-this-line
(car '(a b))
```

prints an error for the first line and then evaluates the second line.

Parse errors are also reported to standard error, but they terminate the program. End-of-file exits normally.

## Project Structure

- `golisp.go` - Starts the stdin-driven REPL and connects parsing to evaluation.
- `parser.go` - Defines atoms and pairs, lexes input, parses S-expressions, expands quote shorthand, and formats output.
- `eval.go` - Implements evaluation and the core built-in functions.
- `assignment.go` - Implements the global environment `rho`, lookup, assignment, truth values, and predicates.
- `parser_test.go` - Tests parsing and formatting.
- `eval_test.go` - Tests core evaluator behavior and REPL integration.
- `assignment_test.go` - Tests assignment, lookup, predicates, and environment behavior.

## Current Limitations

This is an educational interpreter and intentionally has a small feature set. It does not currently support:

- Strings or string literals
- Numeric runtime values or arithmetic
- Comments
- User-defined functions or `lambda`
- Function-valued bindings or higher-order calls
- Local scope
- Dotted-pair input syntax
- Optional predicates such as `number?` or `not?`
- Standard Lisp semantics for the undefined predicate cases described above


