package deluge

import (
	"github.com/ofux/deluge-dsl/ast"
	"github.com/ofux/deluge-dsl/evaluator"
	"github.com/ofux/deluge-dsl/object"
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
)

type Deluge struct {
	ID             string
	Name           string
	GlobalDuration time.Duration
	Scenarios      map[string]*Scenario

	Status DelugeStatus
}

type delugeBuilder struct {
	name            string
	globalDuration  time.Duration
	scenarioCores   map[string]*scenarioCore
	scenarioConfigs map[string]*scenarioConfig
}

type scenarioCore struct {
	name   string
	script ast.Node
}

type scenarioConfig struct {
	concurrent int
	duration   time.Duration
}

func NewDeluge(ID string, script *ast.Program) *Deluge {
	builder := &delugeBuilder{
		scenarioCores:   make(map[string]*scenarioCore),
		scenarioConfigs: make(map[string]*scenarioConfig),
	}
	ev := evaluator.NewEvaluator()
	if err := ev.AddBuiltin("deluge", builder.CreateDeluge); err != nil {
		log.Fatal(err.Error())
	}
	if err := ev.AddBuiltin("scenario", builder.CreateScenario); err != nil {
		log.Fatal(err.Error())
	}
	ev.Eval(script, object.NewEnvironment())

	dlg := &Deluge{
		ID:             ID,
		Name:           builder.name,
		GlobalDuration: builder.globalDuration,
		Scenarios:      make(map[string]*Scenario),
		Status:         DelugeVirgin,
	}
	for id, sConf := range builder.scenarioConfigs {
		if sCore, ok := builder.scenarioCores[id]; ok {
			dlg.Scenarios[id] = NewScenario(sCore.name, sConf.concurrent, sConf.duration, sCore.script)
		} else {
			log.Fatalf("Scenario '%s' is configured but not defined.", id)
		}
	}
	return dlg
}

func (d *Deluge) Run() {
	log.Infof("Executing %d scenario(s)", len(d.Scenarios))
	start := time.Now()

	d.Status = DelugeInProgress

	var waitg sync.WaitGroup
	for _, scenario := range d.Scenarios {
		waitg.Add(1)
		go func(scenario *Scenario) {
			defer waitg.Done()
			scenario.Run(d.GlobalDuration)
		}(scenario)
	}
	waitg.Wait()

	d.Status = DelugeDoneSuccess
	for _, scenario := range d.Scenarios {
		if scenario.Status == ScenarioDoneError {
			d.Status = DelugeDoneError
			break
		}
	}

	log.Infof("Deluge executed %d scenario(s) in %s", len(d.Scenarios), time.Now().Sub(start).String())
}

func (d *delugeBuilder) CreateDeluge(node ast.Node, args ...object.Object) object.Object {
	if len(args) != 3 {
		log.Fatalf("Expected %d arguments at %s\n", 3, ast.PrintLocation(node))
	}

	name, ok := args[0].(*object.String)
	if !ok {
		log.Fatalf("Expected 1st argument to be a string at %s\n", ast.PrintLocation(node))
	}
	d.name = name.Value

	durationStr, ok := args[1].(*object.String)
	if !ok {
		log.Fatalf("Expected 2nd argument to be a string at %s\n", ast.PrintLocation(node))
	}
	duration, err := time.ParseDuration(durationStr.Value)
	if err != nil {
		log.Fatalf("Expected 2nd argument to be a valid duration at %s. Error: %s\n", ast.PrintLocation(node), err.Error())
	}
	d.globalDuration = duration

	conf, ok := args[2].(*object.Hash)
	if !ok {
		log.Fatalf("Expected 3rd argument to be an object at %s\n", ast.PrintLocation(node))
	}

	for scenarioId, v := range conf.Pairs {
		scenarioConf, ok := v.Value.(*object.Hash)
		if !ok {
			log.Fatalf("Expected scenario configuration to be an object at %s\n", ast.PrintLocation(node))
		}

		concurrentClientsHashPair, ok := scenarioConf.Get("concurrent")
		if !ok {
			log.Fatalf("Expected 'concurrent' value in configuration at %s\n", ast.PrintLocation(node))
		}
		concurrentClients, ok := concurrentClientsHashPair.Value.(*object.Integer)
		if !ok {
			log.Fatalf("Expected 'concurrent' value to be an integer in configuration at %s\n", ast.PrintLocation(node))
		}

		delayHashPair, ok := scenarioConf.Get("delay")
		if !ok {
			log.Fatalf("Expected 'delay' value in configuration at %s\n", ast.PrintLocation(node))
		}
		delayHashStr, ok := delayHashPair.Value.(*object.String)
		if !ok {
			log.Fatalf("Expected 'concurrent' value to be a IterationDuration in configuration at %s\n", ast.PrintLocation(node))
		}
		delayHash, err := time.ParseDuration(delayHashStr.Value)
		if err != nil {
			log.Fatalf("Expected 'concurrent' value to be a IterationDuration in configuration at %s\n", ast.PrintLocation(node))
		}

		_, ok = d.scenarioConfigs[string(scenarioId)]
		if ok {
			log.Fatalf("Scenario '%s' is already configured", scenarioId)
		}

		d.scenarioConfigs[string(scenarioId)] = &scenarioConfig{
			concurrent: int(concurrentClients.Value),
			duration:   delayHash,
		}
	}

	return evaluator.NULL
}

func (d *delugeBuilder) CreateScenario(node ast.Node, args ...object.Object) object.Object {
	if len(args) != 3 {
		log.Fatalf("Expected %d arguments at %s\n", 3, ast.PrintLocation(node))
	}

	scenarioId, ok := args[0].(*object.String)
	if !ok {
		log.Fatalf("Expected 1st argument to be a string at %s\n", ast.PrintLocation(node))
	}

	name, ok := args[1].(*object.String)
	if !ok {
		log.Fatalf("Expected 2nd argument to be a string at %s\n", ast.PrintLocation(node))
	}

	coreFunc, ok := args[2].(*object.Function)
	if !ok {
		log.Fatalf("Expected 3rd argument to be a function at %s\n", ast.PrintLocation(node))
	}

	_, ok = d.scenarioCores[scenarioId.Value]
	if ok {
		log.Fatalf("Scenario '%s' is already defined", scenarioId)
	}

	d.scenarioCores[scenarioId.Value] = &scenarioCore{
		name:   name.Value,
		script: coreFunc.Body,
	}

	return evaluator.NULL
}
