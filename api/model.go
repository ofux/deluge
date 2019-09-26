package api

import (
	"errors"
	"fmt"
	"github.com/ofux/deluge/core"
	"github.com/ofux/deluge/core/reporting"
	"github.com/ofux/deluge/dsl/object"
	"time"
)

type DelugeStatus string

const (
	DelugeVirgin      DelugeStatus = "Virgin"
	DelugeInProgress  DelugeStatus = "InProgress"
	DelugeDoneSuccess DelugeStatus = "DoneSuccess"
	DelugeDoneError   DelugeStatus = "DoneError"
	DelugeInterrupted DelugeStatus = "Interrupted"
)

type ScenarioStatus string

const (
	ScenarioVirgin      ScenarioStatus = "Virgin"
	ScenarioInProgress  ScenarioStatus = "InProgress"
	ScenarioDoneSuccess ScenarioStatus = "DoneSuccess"
	ScenarioDoneError   ScenarioStatus = "DoneError"
	ScenarioInterrupted ScenarioStatus = "Interrupted"
)

type Deluge struct {
	ID             string
	Name           string
	Status         DelugeStatus
	GlobalDuration time.Duration
	Scenarios      map[string]*Scenario
}

type DelugeLite struct {
	ID     string
	Name   string
	Status DelugeStatus
}

type Scenario struct {
	Name              string
	IterationDuration time.Duration
	Status            ScenarioStatus
	Errors            []*object.Error
	Report            reporting.Report
}

func mapDeluge(d *core.Deluge) *Deluge {
	d.Mutex.Lock()
	dDTO := &Deluge{
		ID:             d.ID,
		Name:           d.Name,
		GlobalDuration: d.GlobalDuration,
		Status:         mapDelugeStatus(d.Status),
		Scenarios:      make(map[string]*Scenario),
	}
	d.Mutex.Unlock()
	for scID, sc := range d.Scenarios {
		sc.Mutex.Lock()
		dDTO.Scenarios[scID] = mapScenario(sc)
		sc.Mutex.Unlock()
	}
	return dDTO
}

func mapDelugeLite(d *core.Deluge) *DelugeLite {
	d.Mutex.Lock()
	dDTO := &DelugeLite{
		ID:     d.ID,
		Name:   d.Name,
		Status: mapDelugeStatus(d.Status),
	}
	d.Mutex.Unlock()
	return dDTO
}

func mapScenario(sc *core.RunnableScenario) *Scenario {
	return &Scenario{
		Name:              sc.GetScenarioDefinition().Name,
		IterationDuration: sc.IterationDuration,
		Errors:            sc.Errors,
		Report:            sc.Report,
		Status:            mapScenarioStatus(sc.Status),
	}
}

func mapScenarioStatus(st core.ScenarioStatus) ScenarioStatus {
	switch st {
	case core.ScenarioVirgin:
		return ScenarioVirgin
	case core.ScenarioInProgress:
		return ScenarioInProgress
	case core.ScenarioDoneSuccess:
		return ScenarioDoneSuccess
	case core.ScenarioDoneError:
		return ScenarioDoneError
	case core.ScenarioInterrupted:
		return ScenarioInterrupted
	}
	panic(errors.New(fmt.Sprintf("Invalid scenario status %d", st)))
}

func mapDelugeStatus(st core.DelugeStatus) DelugeStatus {
	switch st {
	case core.DelugeVirgin:
		return DelugeVirgin
	case core.DelugeInProgress:
		return DelugeInProgress
	case core.DelugeDoneSuccess:
		return DelugeDoneSuccess
	case core.DelugeDoneError:
		return DelugeDoneError
	case core.DelugeInterrupted:
		return DelugeInterrupted
	}
	panic(errors.New(fmt.Sprintf("Invalid deluge status %d", st)))
}
