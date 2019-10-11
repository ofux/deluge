package core

import (
	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/evaluator"
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/object"
	"github.com/ofux/deluge/dsl/parser"
	"github.com/ofux/deluge/repov2"
	"github.com/pkg/errors"
	"log"
	"time"
)

type DelugeDefinition struct {
	ID             string
	Name           string
	Script         string
	GlobalDuration time.Duration
}

type CompiledDeluge struct {
	definition      *DelugeDefinition
	scenarioConfigs map[string]*scenarioConfig
}

func (c *CompiledDeluge) GetDelugeDefinition() *DelugeDefinition {
	return c.definition
}

func (c *CompiledDeluge) MapToPersistedDeluge() *repov2.PersistedDeluge {
	pd := &repov2.PersistedDeluge{
		ID:             c.definition.ID,
		Name:           c.definition.Name,
		Script:         c.definition.Script,
		GlobalDuration: c.definition.GlobalDuration,
	}
	for scenarioID := range c.scenarioConfigs {
		pd.ScenarioIDs = append(pd.ScenarioIDs, scenarioID)
	}
	return pd
}

func CompileDeluge(script string) (*CompiledDeluge, error) {
	l := lexer.New(script)
	p := parser.New(l)

	program, ok := p.ParseProgram()
	if !ok {
		return nil, p.Errors()
	}

	builder := &delugeBuilder{
		scenarioConfigs: make(map[string]*scenarioConfig),
	}
	ev := evaluator.NewEvaluator()
	if err := ev.AddBuiltin("deluge", builder.dslCreateDeluge); err != nil {
		log.Fatal(err.Error())
	}

	evaluated := ev.Eval(program, object.NewEnvironment())
	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		return nil, errors.New(evaluated.Inspect())
	}

	return &CompiledDeluge{
		definition: &DelugeDefinition{
			ID:             builder.ID,
			Name:           builder.name,
			Script:         script,
			GlobalDuration: builder.globalDuration,
		},
		scenarioConfigs: builder.scenarioConfigs,
	}, nil
}

type delugeBuilder struct {
	visited         bool
	ID              string
	name            string
	globalDuration  time.Duration
	scenarioConfigs map[string]*scenarioConfig
}

type scenarioConfig struct {
	concurrent        int
	iterationDuration time.Duration
	args              *object.Hash
}

func (d *delugeBuilder) dslCreateDeluge(node ast.Node, args ...object.Object) object.Object {
	if d.visited {
		return evaluator.NewError(node, "Expected only one deluge definition at %s\n", ast.PrintLocation(node))
	}
	d.visited = true

	if len(args) != 4 {
		return evaluator.NewError(node, "Expected %d arguments at %s\n", 4, ast.PrintLocation(node))
	}

	delugeId, ok := args[0].(*object.String)
	if !ok || len(delugeId.Value) < 3 {
		return evaluator.NewError(node, "Expected 1st argument to be a string with at least 3 characters at %s\n", ast.PrintLocation(node))
	}
	d.ID = delugeId.Value

	name, ok := args[1].(*object.String)
	if !ok {
		return evaluator.NewError(node, "Expected 2nd argument to be a string at %s\n", ast.PrintLocation(node))
	}
	d.name = name.Value

	durationStr, ok := args[2].(*object.String)
	if !ok {
		return evaluator.NewError(node, "Expected 3rd argument to be a string at %s\n", ast.PrintLocation(node))
	}
	duration, err := time.ParseDuration(durationStr.Value)
	if err != nil {
		return evaluator.NewError(node, "Expected 3rd argument to be a valid duration at %s. Error: %s\n", ast.PrintLocation(node), err.Error())
	}
	d.globalDuration = duration

	conf, ok := args[3].(*object.Hash)
	if !ok {
		return evaluator.NewError(node, "Expected 4th argument to be an object at %s\n", ast.PrintLocation(node))
	}

	for scenarioId, v := range conf.Pairs {
		scenarioConf, ok := v.(*object.Hash)
		if !ok {
			return evaluator.NewError(node, "Expected scenario configuration to be an object at %s\n", ast.PrintLocation(node))
		}

		concurrentClientsHashValue, ok := scenarioConf.Get("concurrent")
		if !ok {
			return evaluator.NewError(node, "Expected 'concurrent' value in configuration at %s\n", ast.PrintLocation(node))
		}
		concurrentClients, ok := concurrentClientsHashValue.(*object.Integer)
		if !ok {
			return evaluator.NewError(node, "Expected 'concurrent' value to be an integer in configuration at %s\n", ast.PrintLocation(node))
		}

		delayHashValue, ok := scenarioConf.Get("delay")
		if !ok {
			return evaluator.NewError(node, "Expected 'delay' value in configuration at %s\n", ast.PrintLocation(node))
		}
		delayHashStr, ok := delayHashValue.(*object.String)
		if !ok {
			return evaluator.NewError(node, "Expected 'delay' value to be a valid duration in configuration at %s\n", ast.PrintLocation(node))
		}
		delayHash, err := time.ParseDuration(delayHashStr.Value)
		if err != nil {
			return evaluator.NewError(node, "Expected 'delay' value to be a valid duration in configuration at %s\n", ast.PrintLocation(node))
		}

		var argsHash *object.Hash
		if argsHashValue, ok := scenarioConf.Get("args"); ok {
			argsHash, ok = argsHashValue.(*object.Hash)
			if !ok {
				return evaluator.NewError(node, "Expected 'args' to be an object at %s\n", ast.PrintLocation(node))
			}
			argsHash.IsImmutable = true
		} else {
			argsHash = &object.Hash{
				Pairs:       map[object.HashKey]object.Object{},
				IsImmutable: true,
			}
		}

		_, ok = d.scenarioConfigs[string(scenarioId)]
		if ok {
			return evaluator.NewError(node, "Scenario '%v' is already configured", scenarioId)
		}

		d.scenarioConfigs[string(scenarioId)] = &scenarioConfig{
			concurrent:        int(concurrentClients.Value),
			iterationDuration: delayHash,
			args:              argsHash,
		}
	}

	return evaluator.NULL
}
