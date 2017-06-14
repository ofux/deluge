package dto

import (
	"errors"
	"fmt"
	"github.com/ofux/deluge-dsl/object"
	"github.com/ofux/deluge/deluge"
	"github.com/ofux/deluge/deluge/reporting"
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

func MapDeluge(d *deluge.Deluge) *Deluge {
	dDTO := &Deluge{
		ID:             d.ID,
		Name:           d.Name,
		GlobalDuration: d.GlobalDuration,
		Status:         MapDelugeStatus(d.Status),
		Scenarios:      make(map[string]*Scenario),
	}
	for k, v := range d.Scenarios {
		dDTO.Scenarios[k] = MapScenario(v)
	}
	return dDTO
}

func MapDelugeLite(d *deluge.Deluge) *DelugeLite {
	dDTO := &DelugeLite{
		ID:     d.ID,
		Name:   d.Name,
		Status: MapDelugeStatus(d.Status),
	}
	return dDTO
}

func MapScenario(sc *deluge.Scenario) *Scenario {
	return &Scenario{
		Name:              sc.Name,
		IterationDuration: sc.IterationDuration,
		Errors:            sc.Errors,
		Report:            sc.Report,
		Status:            MapScenarioStatus(sc.Status),
	}
}

func MapScenarioStatus(st deluge.ScenarioStatus) ScenarioStatus {
	switch st {
	case deluge.ScenarioVirgin:
		return ScenarioVirgin
	case deluge.ScenarioInProgress:
		return ScenarioInProgress
	case deluge.ScenarioDoneSuccess:
		return ScenarioDoneSuccess
	case deluge.ScenarioDoneError:
		return ScenarioDoneError
	}
	panic(errors.New(fmt.Sprintf("Invalid scenario status %d", st)))
}

func MapDelugeStatus(st deluge.DelugeStatus) DelugeStatus {
	switch st {
	case deluge.DelugeVirgin:
		return DelugeVirgin
	case deluge.DelugeInProgress:
		return DelugeInProgress
	case deluge.DelugeDoneSuccess:
		return DelugeDoneSuccess
	case deluge.DelugeDoneError:
		return DelugeDoneError
	}
	panic(errors.New(fmt.Sprintf("Invalid deluge status %d", st)))
}
