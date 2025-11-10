package deluge

import (
	"time"

	"github.com/ofux/deluge/dsl/ast"
)

type Scenario struct {
	Name               string
	TotalSimUsersCalls int64
}

func NewScenario(name string, concurrent int, delay time.Duration, totalDuration time.Duration, script ast.Node) *Scenario {
	return nil
}

func (sc *Scenario) Run(duration time.Duration) {
}
