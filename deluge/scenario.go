package deluge

import (
	"encoding/json"
	"github.com/ofux/deluge-dsl/ast"
	"github.com/ofux/deluge/deluge/recording"
	"github.com/ofux/deluge/deluge/reporting"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Scenario struct {
	Name              string
	simUsers          []*SimUser
	script            ast.Node
	iterationDuration time.Duration
	httpRecorder      *recording.HTTPRecorder
	log               *log.Entry
}

func NewScenario(name string, concurrent int, duration time.Duration, script ast.Node) *Scenario {
	s := &Scenario{
		Name:              name,
		script:            script,
		iterationDuration: duration,
		simUsers:          make([]*SimUser, concurrent),

		httpRecorder: recording.NewHTTPRecorder(concurrent),
		log: log.New().WithFields(log.Fields{
			"scenario": name,
		}),
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

	for _, su := range sc.simUsers {
		waitg.Add(1)
		go func(su *SimUser) {
			defer waitg.Done()
			defer func() {
				atomic.AddUint64(&userCount, 1)
			}()
			ticker := time.NewTicker(sc.iterationDuration)
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

	log.Infof("Scenario executed in %s simulating %d users for %d executions", time.Now().Sub(start).String(), userCount, userExecCount)
	reporter := &reporting.HTTPReporter{}
	if report, err := reporter.Report(sc.httpRecorder); err == nil {
		if jsonReport, err := json.MarshalIndent(report, "", "    "); err == nil {
			err = ioutil.WriteFile("output.json", jsonReport, 0644)
			if err != nil {
				log.Error(err)
			}
		} else {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}
}
