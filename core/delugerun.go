package core

import (
	"github.com/ofux/deluge/repov2"
	"github.com/pkg/errors"
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

type RunnableDeluge struct {
	compiledDeluge *CompiledDeluge
	Scenarios      map[string]*RunnableScenario

	Status    DelugeStatus
	Mutex     *sync.Mutex
	interrupt chan struct{}
}

// GetDelugeDefinition returns a copy of the deluge definition
func (d *RunnableDeluge) GetDelugeDefinition() DelugeDefinition {
	return *d.compiledDeluge.GetDelugeDefinition()
}

func (d *RunnableDeluge) GetGlobalDuration() time.Duration {
	return d.compiledDeluge.globalDuration
}

func NewRunnableDeluge(delugeID string) (*RunnableDeluge, error) {
	persistedDeluge, ok := repov2.Instance.GetDeluge(delugeID)
	if !ok {
		return nil, errors.Errorf("deluge with ID '%s' does not exist", delugeID)
	}
	compiledDeluge, err := CompileDeluge(persistedDeluge.Script)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to recompile deluge with ID '%s'", delugeID)
	}

	dlg := &RunnableDeluge{
		compiledDeluge: compiledDeluge,
		Scenarios:      make(map[string]*RunnableScenario),
		Status:         DelugeVirgin,
		Mutex:          &sync.Mutex{},
		interrupt:      make(chan struct{}),
	}
	for id, sConf := range compiledDeluge.scenarioConfigs {
		if persistedScenario, ok := repov2.Instance.GetScenario(id); ok {
			compiledScenario, err := CompileScenario(persistedScenario.Script)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to recompile scenario %s", id)
			}
			dlg.Scenarios[id] = newRunnableScenario(
				compiledScenario,
				sConf.concurrent,
				sConf.iterationDuration,
				sConf.args,
				log.New().WithField("deluge", dlg.GetDelugeDefinition().Name),
			)
		} else {
			return nil, errors.Errorf("scenario '%s' is configured but not defined", id)
		}
	}
	return dlg, nil
}

// Run runs the deluge asynchronously. It returns a channel that will be closed once the execution is finished.
func (d *RunnableDeluge) Run() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		d.run()
	}()
	return done
}

func (d *RunnableDeluge) run() {
	log.Infof("Executing %d scenario(s)", len(d.Scenarios))
	start := time.Now()

	d.Mutex.Lock()
	if d.Status != DelugeVirgin {
		log.Warnf("Cannot run a deluge %s with status %d", d.GetDelugeDefinition().ID, d.Status)
		return
	}
	d.Status = DelugeInProgress
	d.Mutex.Unlock()

	var waitg sync.WaitGroup
	for _, scenario := range d.Scenarios {
		waitg.Add(1)
		go func(scenario *RunnableScenario) {
			defer waitg.Done()
			scenario.run(d.compiledDeluge.globalDuration, d.interrupt)
		}(scenario)
	}
	waitg.Wait()

	d.Mutex.Lock()
	d.end()
	d.Mutex.Unlock()

	log.Infof("Deluge executed %d scenario(s) in %s", len(d.Scenarios), time.Now().Sub(start).String())
}

func (d *RunnableDeluge) end() {
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

func (d *RunnableDeluge) Interrupt() {
	d.Mutex.Lock()
	if d.Status == DelugeVirgin || d.Status == DelugeInProgress {
		d.Status = DelugeInterrupted
		d.Mutex.Unlock()
		close(d.interrupt)
	} else {
		d.Mutex.Unlock()
	}
}
