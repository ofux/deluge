package core

import (
	"errors"
	"fmt"
	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/evaluator"
	"github.com/ofux/deluge/dsl/object"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type DelugeStatus int

const (
	DelugeVirgin DelugeStatus = iota
	DelugeInProgress
	DelugeDoneSuccess
	DelugeDoneError
	DelugeInterrupted
)

type Deluge struct {
	ID             string
	Name           string
	GlobalDuration time.Duration
	Scenarios      map[string]*Scenario

	Status    DelugeStatus
	Mutex     *sync.Mutex
	interrupt chan struct{}
}

type delugeBuilder struct {
	name            string
	globalDuration  time.Duration
	scenarioCores   map[string]*scenarioCore
	scenarioConfigs map[string]*scenarioConfig
}

type scenarioCore struct {
	name         string
	script       ast.Node
	scriptParams []*ast.Identifier
}

type scenarioConfig struct {
	concurrent        int
	iterationDuration time.Duration
}

func NewDeluge(ID string, script *ast.Program) (*Deluge, error) {
	builder := &delugeBuilder{
		scenarioCores:   make(map[string]*scenarioCore),
		scenarioConfigs: make(map[string]*scenarioConfig),
	}
	ev := evaluator.NewEvaluator()
	if err := ev.AddBuiltin("deluge", builder.dslCreateDeluge); err != nil {
		log.Fatal(err.Error())
	}
	if err := ev.AddBuiltin("scenario", builder.dslCreateScenario); err != nil {
		log.Fatal(err.Error())
	}

	evaluated := ev.Eval(script, object.NewEnvironment())
	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		return nil, errors.New(evaluated.Inspect())
	}

	dlg := &Deluge{
		ID:             ID,
		Name:           builder.name,
		GlobalDuration: builder.globalDuration,
		Scenarios:      make(map[string]*Scenario),
		Status:         DelugeVirgin,
		Mutex:          &sync.Mutex{},
		interrupt:      make(chan struct{}),
	}
	for id, sConf := range builder.scenarioConfigs {
		if sCore, ok := builder.scenarioCores[id]; ok {
			dlg.Scenarios[id] = newScenario(
				sCore.name,
				sConf.concurrent,
				sConf.iterationDuration,
				sCore.script,
				sCore.scriptParams,
				nil,
				log.New().WithField("deluge", dlg.Name),
			)
		} else {
			return nil, errors.New(fmt.Sprintf("Scenario '%s' is configured but not defined.", id))
		}
	}
	return dlg, nil
}

// Run runs the deluge asynchronously. It returns a channel that will be closed once the execution is finished.
func (d *Deluge) Run() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		d.run()
	}()
	return done
}

func (d *Deluge) run() {
	log.Infof("Executing %d scenario(s)", len(d.Scenarios))
	start := time.Now()

	d.Mutex.Lock()
	if d.Status != DelugeVirgin {
		log.Warnf("Cannot run a deluge %s with status %d", d.ID, d.Status)
		return
	}
	d.Status = DelugeInProgress
	d.Mutex.Unlock()

	var waitg sync.WaitGroup
	for _, scenario := range d.Scenarios {
		waitg.Add(1)
		go func(scenario *Scenario) {
			defer waitg.Done()
			scenario.run(d.GlobalDuration, d.interrupt)
		}(scenario)
	}
	waitg.Wait()

	d.Mutex.Lock()
	d.end()
	d.Mutex.Unlock()

	log.Infof("Deluge executed %d scenario(s) in %s", len(d.Scenarios), time.Now().Sub(start).String())
}

func (d *Deluge) end() {
	if d.Status == DelugeInProgress {
		d.Status = DelugeDoneSuccess
		for _, scenario := range d.Scenarios {
			if scenario.Status == ScenarioDoneError {
				d.Status = DelugeDoneError
				break
			}
		}
	}
}

func (d *Deluge) Interrupt() {
	d.Mutex.Lock()
	if d.Status == DelugeVirgin || d.Status == DelugeInProgress {
		d.Status = DelugeInterrupted
		d.Mutex.Unlock()
		close(d.interrupt)
	} else {
		d.Mutex.Unlock()
	}
}

func (d *delugeBuilder) dslCreateDeluge(node ast.Node, args ...object.Object) object.Object {
	if len(args) != 3 {
		return evaluator.NewError(node, "Expected %d arguments at %s\n", 3, ast.PrintLocation(node))
	}

	name, ok := args[0].(*object.String)
	if !ok {
		return evaluator.NewError(node, "Expected 1st argument to be a string at %s\n", ast.PrintLocation(node))
	}
	d.name = name.Value

	durationStr, ok := args[1].(*object.String)
	if !ok {
		return evaluator.NewError(node, "Expected 2nd argument to be a string at %s\n", ast.PrintLocation(node))
	}
	duration, err := time.ParseDuration(durationStr.Value)
	if err != nil {
		return evaluator.NewError(node, "Expected 2nd argument to be a valid duration at %s. Error: %s\n", ast.PrintLocation(node), err.Error())
	}
	d.globalDuration = duration

	conf, ok := args[2].(*object.Hash)
	if !ok {
		return evaluator.NewError(node, "Expected 3rd argument to be an object at %s\n", ast.PrintLocation(node))
	}

	for scenarioId, v := range conf.Pairs {
		scenarioConf, ok := v.Value.(*object.Hash)
		if !ok {
			return evaluator.NewError(node, "Expected scenario configuration to be an object at %s\n", ast.PrintLocation(node))
		}

		concurrentClientsHashPair, ok := scenarioConf.Get("concurrent")
		if !ok {
			return evaluator.NewError(node, "Expected 'concurrent' value in configuration at %s\n", ast.PrintLocation(node))
		}
		concurrentClients, ok := concurrentClientsHashPair.Value.(*object.Integer)
		if !ok {
			return evaluator.NewError(node, "Expected 'concurrent' value to be an integer in configuration at %s\n", ast.PrintLocation(node))
		}

		delayHashPair, ok := scenarioConf.Get("delay")
		if !ok {
			return evaluator.NewError(node, "Expected 'delay' value in configuration at %s\n", ast.PrintLocation(node))
		}
		delayHashStr, ok := delayHashPair.Value.(*object.String)
		if !ok {
			return evaluator.NewError(node, "Expected 'delay' value to be a valid duration in configuration at %s\n", ast.PrintLocation(node))
		}
		delayHash, err := time.ParseDuration(delayHashStr.Value)
		if err != nil {
			return evaluator.NewError(node, "Expected 'delay' value to be a valid duration in configuration at %s\n", ast.PrintLocation(node))
		}

		_, ok = d.scenarioConfigs[string(scenarioId)]
		if ok {
			return evaluator.NewError(node, "Scenario '%v' is already configured", scenarioId)
		}

		d.scenarioConfigs[string(scenarioId)] = &scenarioConfig{
			concurrent:        int(concurrentClients.Value),
			iterationDuration: delayHash,
		}
	}

	return evaluator.NULL
}

func (d *delugeBuilder) dslCreateScenario(node ast.Node, args ...object.Object) object.Object {
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

	_, ok = d.scenarioCores[scenarioId.Value]
	if ok {
		return evaluator.NewError(node, "Scenario '%s' is already defined", scenarioId.Value)
	}

	d.scenarioCores[scenarioId.Value] = &scenarioCore{
		name:         name.Value,
		script:       coreFunc.Body,
		scriptParams: coreFunc.Parameters,
	}

	return evaluator.NULL
}
