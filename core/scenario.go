package core

import (
	"github.com/ofux/deluge-dsl/ast"
	"github.com/ofux/deluge-dsl/object"
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/core/reporting"
	log "github.com/sirupsen/logrus"
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

	Status ScenarioStatus
	Errors []*object.Error
	Report reporting.Report
}

func NewScenario(name string, concurrent int, duration time.Duration, script ast.Node) *Scenario {
	s := &Scenario{
		Name:              name,
		script:            script,
		IterationDuration: duration,
		simUsers:          make([]*SimUser, concurrent),

		httpRecorder: recording.NewHTTPRecorder(concurrent),
		log: log.New().WithFields(log.Fields{
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
	var userCount uint64 = 0
	var userExecCount uint64 = 0

	start := time.Now()

	sc.Status = ScenarioInProgress

	for _, su := range sc.simUsers {
		waitg.Add(1)
		go func(su *SimUser) {
			defer waitg.Done()
			defer func() {
				atomic.AddUint64(&userCount, 1)
			}()
			ticker := time.NewTicker(sc.IterationDuration)
			timer := time.NewTimer(globalDuration)

			i := 0
			for {
				if time.Now().Sub(start).Nanoseconds() > globalDuration.Nanoseconds() {
					log.Debugf("Terminate user simulation %s", su.Name)
					return
				}

				log.Debugf("Running user simulation %s", su.Name)
				su.Run(i)

				if su.Status == DoneError {
					return
				}

				select {
				case <-timer.C:
					log.Debugf("Terminate user simulation %s", su.Name)
					return
				case <-ticker.C:
				}
				i++
				atomic.AddUint64(&userExecCount, 1)
			}
		}(su)
	}
	waitg.Wait()
	sc.httpRecorder.Close()

	sc.Status = ScenarioDoneSuccess
	for _, su := range sc.simUsers {
		if su.Status == DoneError {
			sc.Status = ScenarioDoneError
			sc.Errors = append(sc.Errors, su.Error)
		}
	}

	log.Infof("Scenario executed in %s simulating %d users for %d executions", time.Now().Sub(start).String(), userCount, userExecCount)

	reporter := &reporting.HTTPReporter{}
	if report, err := reporter.Report(sc.httpRecorder); err == nil {
		sc.Report = report
	} else {
		log.Error(err)
	}
}
