package dto

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
)

type ScenarioStatus string

const (
	ScenarioVirgin      ScenarioStatus = "Virgin"
	ScenarioInProgress  ScenarioStatus = "InProgress"
	ScenarioDoneSuccess ScenarioStatus = "DoneSuccess"
	ScenarioDoneError   ScenarioStatus = "DoneError"
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

func MapDeluge(d *core.Deluge) *Deluge {
	d.Mutex.Lock()
	dDTO := &Deluge{
		ID:             d.ID,
		Name:           d.Name,
		GlobalDuration: d.GlobalDuration,
		Status:         MapDelugeStatus(d.Status),
		Scenarios:      make(map[string]*Scenario),
	}
	d.Mutex.Unlock()
	for scID, sc := range d.Scenarios {
		sc.Mutex.Lock()
		dDTO.Scenarios[scID] = MapScenario(sc)
		sc.Mutex.Unlock()
	}
	return dDTO
}

func MapDelugeLite(d *core.Deluge) *DelugeLite {
	d.Mutex.Lock()
	dDTO := &DelugeLite{
		ID:     d.ID,
		Name:   d.Name,
		Status: MapDelugeStatus(d.Status),
	}
	d.Mutex.Unlock()
	return dDTO
}

func MapScenario(sc *core.Scenario) *Scenario {
	return &Scenario{
		Name:              sc.Name,
		IterationDuration: sc.IterationDuration,
		Errors:            sc.Errors,
		Report:            sc.Report,
		Status:            MapScenarioStatus(sc.Status),
	}
}

func MapScenarioStatus(st core.ScenarioStatus) ScenarioStatus {
	switch st {
	case core.ScenarioVirgin:
		return ScenarioVirgin
	case core.ScenarioInProgress:
		return ScenarioInProgress
	case core.ScenarioDoneSuccess:
		return ScenarioDoneSuccess
	case core.ScenarioDoneError:
		return ScenarioDoneError
	}
	panic(errors.New(fmt.Sprintf("Invalid scenario status %d", st)))
}

func MapDelugeStatus(st core.DelugeStatus) DelugeStatus {
	switch st {
	case core.DelugeVirgin:
		return DelugeVirgin
	case core.DelugeInProgress:
		return DelugeInProgress
	case core.DelugeDoneSuccess:
		return DelugeDoneSuccess
	case core.DelugeDoneError:
		return DelugeDoneError
	}
	panic(errors.New(fmt.Sprintf("Invalid deluge status %d", st)))
}
