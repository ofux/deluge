package deluge

import (
	"github.com/ofux/deluge-dsl/ast"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

type Scenario struct {
	Name              string
	simUsers          []*SimUser
	script            ast.Node
	iterationDuration time.Duration
	recorder          Recorder
	log               *log.Entry
}

func NewScenario(name string, concurrent int, duration time.Duration, script ast.Node) *Scenario {
	s := &Scenario{
		Name:              name,
		script:            script,
		iterationDuration: duration,
		simUsers:          make([]*SimUser, concurrent),

		recorder: NewRecorder(concurrent),
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

	start := time.Now()

	for _, su := range sc.simUsers {
		waitg.Add(1)
		go func(su *SimUser) {
			defer waitg.Done()
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
			}
		}(su)
	}
	waitg.Wait()
	sc.recorder.Close()

	log.Infof("Scenario executed in %s", time.Now().Sub(start).String())
}
