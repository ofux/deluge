package core

import (
	"github.com/ofux/deluge-dsl/ast"
	"github.com/ofux/deluge-dsl/object"
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/core/reporting"
	log "github.com/sirupsen/logrus"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type ScenarioStatus int

const (
	ScenarioVirgin ScenarioStatus = iota
	ScenarioInProgress
	ScenarioDoneSuccess
	ScenarioDoneError
)

type Scenario struct {
	Name              string
	simUsers          []*SimUser
	script            ast.Node
	IterationDuration time.Duration
	httpRecorder      *recording.HTTPRecorder
	log               *log.Entry

	Status             ScenarioStatus
	Errors             []*object.Error
	Report             reporting.Report
	EffectiveUserCount uint64
	EffectiveExecCount uint64
}

func NewScenario(name string, concurrent int, iterationDuration time.Duration, script ast.Node, logEntry *log.Entry) *Scenario {
	s := &Scenario{
		Name:              name,
		script:            script,
		IterationDuration: iterationDuration,
		simUsers:          make([]*SimUser, concurrent),

		httpRecorder: recording.NewHTTPRecorder(),
		log: logEntry.WithFields(log.Fields{
			"scenario": name,
		}),

		Status: ScenarioVirgin,
		Errors: make([]*object.Error, 0),
	}

	for i := 0; i < concurrent; i++ {
		s.simUsers[i] = NewSimUser(strconv.Itoa(i), s)
	}

	return s
}

func (sc *Scenario) Run(globalDuration time.Duration) {
	var waitg sync.WaitGroup

	start := time.Now()
	endTime := start.Add(globalDuration)

	sc.Status = ScenarioInProgress

	for _, su := range sc.simUsers {
		waitg.Add(1)
		go func(su *SimUser) {
			defer waitg.Done()
			defer func() {
				atomic.AddUint64(&sc.EffectiveUserCount, 1)
			}()

			i := 0
			for time.Now().Before(endTime) {
				iterationEndTime := time.Now().Add(sc.IterationDuration)

				sc.log.Debugf("Running user simulation %s", su.Name)
				su.Run(i)
				i++
				atomic.AddUint64(&sc.EffectiveExecCount, 1)

				if su.Status == DoneError {
					return
				}

				// Check if we're going to reach endTime
				if iterationEndTime.Before(endTime) {
					// Wait till the end of iteration as defined in scenario configuration
					if time.Now().Before(iterationEndTime) {
						time.Sleep(time.Until(iterationEndTime))
					} else {
						// In case we already reached iterationEndTime, we do not sleep, but we add a schedule point
						// because we cannot assume there is one in the simulation execution itself.
						runtime.Gosched()
					}
				} else {
					sc.log.Debugf("Terminate user simulation %s", su.Name)
					return
				}
			}
		}(su)
	}
	waitg.Wait()

	sc.end()

	sc.log.Infof("Scenario executed in %s simulating %d users for %d executions", time.Now().Sub(start).String(), sc.EffectiveUserCount, sc.EffectiveExecCount)
}

func (sc *Scenario) end() {

	sc.httpRecorder.Close()
	reporter := &reporting.HTTPReporter{}
	if report, err := reporter.Report(sc.httpRecorder); err == nil {
		sc.Report = report
	} else {
		sc.log.Error(err)
	}

	sc.Status = ScenarioDoneSuccess
	for _, su := range sc.simUsers {
		if su.Status == DoneError {
			sc.Status = ScenarioDoneError
			sc.Errors = append(sc.Errors, su.Error)
		}
	}

	// we don't need simulated users anymore
	sc.simUsers = nil
}
