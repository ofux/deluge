package repl

import (
	"bufio"
	"fmt"
	"github.com/ofux/deluge/dsl/evaluator"
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/object"
	"github.com/ofux/deluge/dsl/parser"
	"io"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program, ok := p.ParseProgram()
		if !ok {
			printParserErrors(out, p.Errors())
			continue
		}

		ev := evaluator.NewEvaluator()

		evaluated := ev.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func printParserErrors(out io.Writer, errors []parser.ParseError) {
	io.WriteString(out, "Syntax error:\n")
	for _, err := range errors {
		io.WriteString(out, fmt.Sprintf("\t%s (line %d, col %d)\n", err.Message, err.Line, err.Column))
	}
}
