package core

import (
	"errors"
	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/evaluator"
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/object"
	"github.com/ofux/deluge/dsl/parser"
	"log"
)

type ScenarioDefinition struct {
	ID     string
	Name   string
	Script string
}

type CompiledScenario struct {
	scenario     *ScenarioDefinition
	script       ast.Node
	scriptParams []*ast.Identifier
}

func (c *CompiledScenario) GetScenarioDefinition() *ScenarioDefinition {
	return c.scenario
}

func CompileScenario(script string) (*CompiledScenario, error) {
	l := lexer.New(script)
	p := parser.New(l)

	program, ok := p.ParseProgram()
	if !ok {
		return nil, p.Errors()
	}

	builder := &scenarioBuilder{}
	ev := evaluator.NewEvaluator()
	if err := ev.AddBuiltin("scenario", builder.dslCreateScenario); err != nil {
		log.Fatal(err.Error())
	}

	evaluated := ev.Eval(program, object.NewEnvironment())
	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		return nil, errors.New(evaluated.Inspect())
	}

	return &CompiledScenario{
		scenario: &ScenarioDefinition{
			ID:     builder.ID,
			Name:   builder.name,
			Script: script,
		},
		script:       builder.script,
		scriptParams: builder.scriptParams,
	}, nil
}

type scenarioBuilder struct {
	ID           string
	name         string
	script       ast.Node
	scriptParams []*ast.Identifier
}

func (d *scenarioBuilder) dslCreateScenario(node ast.Node, args ...object.Object) object.Object {
	if len(args) != 3 {
		return evaluator.NewError(node, "Expected %d arguments at %s\n", 3, ast.PrintLocation(node))
	}

	scenarioId, ok := args[0].(*object.String)
	if !ok {
		return evaluator.NewError(node, "Expected 1st argument to be a string at %s\n", ast.PrintLocation(node))
	}

	name, ok := args[1].(*object.String)
	if !ok {
		return evaluator.NewError(node, "Expected 2nd argument to be a string at %s\n", ast.PrintLocation(node))
	}

	coreFunc, ok := args[2].(*object.Function)
	if !ok {
		return evaluator.NewError(node, "Expected 3rd argument to be a function at %s\n", ast.PrintLocation(node))
	}

	d.ID = scenarioId.Value
	d.name = name.Value
	d.script = coreFunc.Body
	d.scriptParams = coreFunc.Parameters

	return evaluator.NULL
}
