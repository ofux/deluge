package core

import (
	"errors"
	"fmt"
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/core/status"
	"github.com/ofux/deluge/dsl/object"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type RunnableScenario struct {
	compiledScenario  *CompiledScenario
	simUsers          []*simUser
	scriptArgs        *object.Hash
	IterationDuration time.Duration
	httpRecorder      *recording.HTTPRecorder
	log               *log.Entry

	Status             status.ScenarioStatus
	Errors             []*object.Error
	Records            *recording.HTTPRecordsOverTime
	EffectiveUserCount uint64
	EffectiveExecCount uint64
	Mutex              *sync.Mutex
}

func newRunnableScenario(
	compiledScenario *CompiledScenario,
	concurrent int,
	iterationDuration time.Duration,
	scriptArgs *object.Hash,
	logEntry *log.Entry,
) *RunnableScenario {
	s := &RunnableScenario{
		compiledScenario:  compiledScenario,
		scriptArgs:        scriptArgs,
		IterationDuration: iterationDuration,
		simUsers:          make([]*simUser, concurrent),

		httpRecorder: recording.NewHTTPRecorder(1),
		log: logEntry.WithFields(log.Fields{
			"scenario": compiledScenario.scenario.ID,
		}),

		Status: status.ScenarioVirgin,
		Errors: make([]*object.Error, 0),

		Mutex: &sync.Mutex{},
	}

	for i := 0; i < concurrent; i++ {
		s.simUsers[i] = newSimUser(strconv.Itoa(i), s)
	}

	return s
}

// GetScenarioDefinition returns a copy of the scenario definition
func (sc *RunnableScenario) GetScenarioDefinition() ScenarioDefinition {
	return *sc.compiledScenario.scenario
}

func (sc *RunnableScenario) run(globalDuration time.Duration, interrupt chan struct{}) {
	var waitg sync.WaitGroup

	start := time.Now()
	endTime := start.Add(globalDuration)

	sc.Mutex.Lock()
	if sc.Status != status.ScenarioVirgin {
		panic(errors.New(fmt.Sprintf("Cannot run a scenario with status %d", sc.Status)))
	}
	sc.Status = status.ScenarioInProgress
	sc.Mutex.Unlock()

	for _, su := range sc.simUsers {
		waitg.Add(1)
		go func(su *simUser) {
			defer waitg.Done()
			sc.runSimUser(su, endTime, interrupt)
		}(su)
	}
	waitg.Wait()

	sc.Mutex.Lock()
	sc.end()
	sc.Mutex.Unlock()

	sc.log.Infof("Scenario executed in %s simulating %d users for %d executions", time.Now().Sub(start).String(), sc.EffectiveUserCount, sc.EffectiveExecCount)
}

func (sc *RunnableScenario) runSimUser(su *simUser, endTime time.Time, interrupt chan struct{}) {
	defer func() {
		atomic.AddUint64(&sc.EffectiveUserCount, 1)
	}()

	i := 0
	for time.Now().Before(endTime) {
		select {
		case <-interrupt:
			su.status = UserInterrupted
			sc.log.Debugf("Terminate user simulation %s because of interrupt signal.", su.name)
			return
		default:
			iterationEndTime := time.Now().Add(sc.IterationDuration)

			sc.log.Debugf("Running user simulation %s", su.name)
			su.run(i)
			i++
			atomic.AddUint64(&sc.EffectiveExecCount, 1)

			if su.status == UserDoneError {
				sc.log.Debugf("Terminate user simulation %s because an error occurred.", su.name)
				return
			}

			// Check if we're going to reach endTime
			if iterationEndTime.Before(endTime) {
				// Wait till the end of iteration as defined in scenario configuration
				if time.Now().Before(iterationEndTime) {
					time.Sleep(time.Until(iterationEndTime))
				}
			} else {
				sc.log.Debugf("Terminate user simulation %s.", su.name)
				return
			}
		}
	}
}

func (sc *RunnableScenario) end() {

	sc.httpRecorder.Close()
	if records, err := sc.httpRecorder.GetRecords(); err == nil {
		sc.Records = records
	} else {
		sc.Status = status.ScenarioDoneError
		sc.log.Error(err)
	}

	sc.Status = status.ScenarioDoneSuccess
	for _, su := range sc.simUsers {
		if su.status == UserDoneError {
			sc.Status = status.ScenarioDoneError
			sc.Errors = append(sc.Errors, su.execError)
		} else if su.status == UserInterrupted && sc.Status != status.ScenarioDoneError {
			sc.Status = status.ScenarioInterrupted
		}
	}

	// we don't need simulated users anymore
	sc.simUsers = nil
}
