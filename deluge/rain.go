package deluge

import (
	"github.com/robertkrimen/otto"
	"sync"
)

type Rain struct {
	waterdrops []*WaterDrop
}

func NewRain(script *otto.Script, count int) *Rain {
	r := &Rain{
		waterdrops: make([]*WaterDrop, count),
	}

	for i := range r.waterdrops {
		r.waterdrops[i] = NewWaterDrop(script)
	}

	return r
}

func (r *Rain) Run() {
	var waitg sync.WaitGroup

	for _, wd := range r.waterdrops {
		waitg.Add(1)
		go func(wd *WaterDrop) {
			defer waitg.Done()
			wd.Run()
		}(wd)
	}
	waitg.Wait()
}