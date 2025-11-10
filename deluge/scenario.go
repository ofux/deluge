package deluge

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ofux/deluge/dsl/ast"
	log "github.com/sirupsen/logrus"
)

type Scenario struct {
	Name               string
	simUsers           []*SimUser
	script             ast.Node
	delay              time.Duration
	totalDuration      time.Duration
	TotalSimUsersCalls int64
}

func NewScenario(name string, concurrent int, delay time.Duration, totalDuration time.Duration, script ast.Node) *Scenario {
	s := &Scenario{
		Name:          name,
		simUsers:      make([]*SimUser, concurrent),
		delay:         delay,
		totalDuration: totalDuration,
		script:        script,
	}

	for i := 0; i < concurrent; i++ {
		s.simUsers[i] = NewSimUser(strconv.Itoa(i), s.script)
	}

	return s
}

func (sc *Scenario) Run(duration time.Duration) {
	var waitg sync.WaitGroup
	var counterWg sync.WaitGroup

	simUserCallCounter := sc.countSimUserCalls(&counterWg)
	defer func() {
		close(simUserCallCounter)
		counterWg.Wait() // Ensure counter goroutine finishes
	}()

	start := time.Now()

	for _, su := range sc.simUsers {
		waitg.Add(1)
		go func(su *SimUser) {
			defer waitg.Done()
			ticker := time.NewTicker(sc.delay)
			timer := time.NewTimer(duration)

			for {
				if time.Now().Sub(start).Nanoseconds() > duration.Nanoseconds() {
					log.Debugf("Terminate user simulation %s", su.Name)
					return
				}

				log.Debugf("Running user simulation %s", su.Name)
				simUserCallCounter <- 1
				su.Run()

				if su.Status == DoneError {
					return
				}

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
}

func (sc *Scenario) countSimUserCalls(wg *sync.WaitGroup) chan int8 {
	simCallCounter := make(chan int8, 10000)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			_, ok := <-simCallCounter
			if ok {
				atomic.AddInt64(&sc.TotalSimUsersCalls, 1)
			} else {
				return
			}
		}
	}()
	return simCallCounter
}
