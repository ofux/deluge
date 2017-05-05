package deluge

import (
	"github.com/ofux/deluge-dsl/ast"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

type Rain struct {
	Name                 string
	waterdrops           []*WaterDrop
	callsTickerDuration  time.Duration
	TotalWaterDropsCalls int64
}

func NewRain(name string, script *ast.Program, concurrentClients int, callsTickerDuration time.Duration) *Rain {
	r := &Rain{
		Name:                name,
		waterdrops:          make([]*WaterDrop, concurrentClients),
		callsTickerDuration: callsTickerDuration,
	}

	for i := range r.waterdrops {
		r.waterdrops[i] = NewWaterDrop(strconv.Itoa(i), script)
	}

	return r
}

func (r *Rain) Run(duration time.Duration) {
	var waitg sync.WaitGroup

	waterDropCallCounter := r.countWaterDropCalls()
	defer close(waterDropCallCounter)

	start := time.Now()

	for _, wd := range r.waterdrops {
		waitg.Add(1)
		go func(wd *WaterDrop) {
			defer waitg.Done()
			ticker := time.NewTicker(r.callsTickerDuration)
			timer := time.NewTimer(duration)

			for {
				if time.Now().Sub(start).Nanoseconds() > duration.Nanoseconds() {
					log.Debugf("Terminate waterdrop %s", wd.Name)
					return
				}

				log.Debugf("Running waterdrop %s", wd.Name)
				waterDropCallCounter <- 1
				wd.Run()

				select {
				case <-timer.C:
					log.Debugf("Terminate waterdrop %s", wd.Name)
					return
				case <-ticker.C:
				}
			}
		}(wd)
	}
	waitg.Wait()

	log.Infof("Deluge executed %d requests in %s", r.TotalWaterDropsCalls, time.Now().Sub(start).String())
}

func (r *Rain) countWaterDropCalls() chan int8 {
	waterDropCallCounter := make(chan int8, 10000)
	go func() {
		for {
			_, ok := <-waterDropCallCounter
			if ok {
				r.TotalWaterDropsCalls++
			} else {
				return
			}
		}
	}()
	return waterDropCallCounter
}
