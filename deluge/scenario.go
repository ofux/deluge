package deluge

import (
	"github.com/ofux/deluge-dsl/ast"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

type Scenario struct {
	Name               string
	simUsers           []*SimUser
	script             ast.Node
	duration           time.Duration
	TotalSimUsersCalls int64
}

func NewScenario(name string, concurrent int, duration time.Duration, script ast.Node) *Scenario {
	s := &Scenario{
		Name:     name,
		simUsers: make([]*SimUser, concurrent),
		duration: duration,
		script:   script,
	}

	for i := 0; i < concurrent; i++ {
		s.simUsers[i] = NewSimUser(strconv.Itoa(i), s.script)
	}

	return s
}

func (sc *Scenario) Run(duration time.Duration) {
	var waitg sync.WaitGroup

	simUserCallCounter := sc.countSimUserCalls()
	defer close(simUserCallCounter)

	start := time.Now()

	for _, su := range sc.simUsers {
		waitg.Add(1)
		go func(su *SimUser) {
			defer waitg.Done()
			ticker := time.NewTicker(sc.duration)
			timer := time.NewTimer(duration)

			for {
				if time.Now().Sub(start).Nanoseconds() > duration.Nanoseconds() {
					log.Debugf("Terminate user simulation %s", su.Name)
					return
				}

				log.Debugf("Running user simulation %s", su.Name)
				simUserCallCounter <- 1
				su.Run()

				select {
				case <-timer.C:
					log.Debugf("Terminate user simulation %s", su.Name)
					return
				case <-ticker.C:
				}
			}
		}(su)
	}
	waitg.Wait()

	log.Infof("Scenario executed %d requests in %s", sc.TotalSimUsersCalls, time.Now().Sub(start).String())
}

func (sc *Scenario) countSimUserCalls() chan int8 {
	simCallCounter := make(chan int8, 10000)
	go func() {
		for {
			_, ok := <-simCallCounter
			if ok {
				sc.TotalSimUsersCalls++
			} else {
				return
			}
		}
	}()
	return simCallCounter
}
