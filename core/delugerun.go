package core

import (
	"github.com/ofux/deluge/core/status"
	"github.com/ofux/deluge/repov2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type RunnableDeluge struct {
	compiledDeluge *CompiledDeluge
	Scenarios      map[string]*RunnableScenario

	runStatus      status.DelugeStatus
	runStatusMutex *sync.Mutex
	interrupt      chan struct{}
	statusChange   chan status.DelugeStatus
}

// GetDelugeDefinition returns a copy of the deluge definition
func (d *RunnableDeluge) GetDelugeDefinition() DelugeDefinition {
	return *d.compiledDeluge.GetDelugeDefinition()
}

func (d *RunnableDeluge) GetGlobalDuration() time.Duration {
	return d.compiledDeluge.GetDelugeDefinition().GlobalDuration
}

func (d *RunnableDeluge) OnStatusChangeChan() <-chan status.DelugeStatus {
	return d.statusChange
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
		runStatus:      status.DelugeVirgin,
		runStatusMutex: &sync.Mutex{},
		interrupt:      make(chan struct{}),
		statusChange:   make(chan status.DelugeStatus, 5), // Status cannot change more than 5 times
	}
	dlg.statusChange <- dlg.runStatus
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

	d.runStatusMutex.Lock()
	if d.runStatus != status.DelugeVirgin {
		log.Warnf("Cannot run a deluge %s with status %d", d.GetDelugeDefinition().ID, d.runStatus)
		d.runStatusMutex.Unlock()
		return
	}
	d.runStatus = status.DelugeInProgress
	d.statusChange <- d.runStatus
	d.runStatusMutex.Unlock()

	var waitg sync.WaitGroup
	for _, scenario := range d.Scenarios {
		waitg.Add(1)
		go func(scenario *RunnableScenario) {
			defer waitg.Done()
			scenario.run(d.GetGlobalDuration(), d.interrupt)
		}(scenario)
	}
	waitg.Wait()

	d.end()

	log.Infof("Deluge executed %d scenario(s) in %s", len(d.Scenarios), time.Now().Sub(start).String())
}

func (d *RunnableDeluge) end() {
	d.runStatusMutex.Lock()
	defer d.runStatusMutex.Unlock()

	if d.runStatus == status.DelugeInProgress {
		d.runStatus = status.DelugeDoneSuccess
		for _, scenario := range d.Scenarios {
			if scenario.Status == status.ScenarioDoneError {
				d.runStatus = status.DelugeDoneError
				break
			}
		}
		d.statusChange <- d.runStatus
	}
	close(d.statusChange)
}

func (d *RunnableDeluge) Interrupt() {
	d.runStatusMutex.Lock()
	if d.runStatus == status.DelugeVirgin || d.runStatus == status.DelugeInProgress {
		d.runStatus = status.DelugeInterrupted
		d.statusChange <- d.runStatus
		d.runStatusMutex.Unlock()
		close(d.interrupt)
	} else {
		d.runStatusMutex.Unlock()
	}
}
