package deluge

import (
	"time"

	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/evaluator"
	"github.com/ofux/deluge/dsl/object"
)

type SimUserStatus int

const (
	Virgin SimUserStatus = iota
	InProgress
	DoneSuccess
	DoneError
)

type SimUser struct {
	Name          string
	Status        SimUserStatus
	SleepDuration time.Duration
}

func NewSimUser(name string, script ast.Node) *SimUser {
	return nil
}

func (su *SimUser) Run() {
}

func (su *SimUser) ExecHTTPRequest(node ast.Node, args ...object.Object) object.Object {
	return evaluator.NULL
}
