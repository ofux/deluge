package api

import (
	"errors"
	"fmt"
	"github.com/ofux/deluge/core"
	"github.com/ofux/deluge/core/reporting"
	"github.com/ofux/deluge/dsl/object"
	"time"
)

type JobStatus string

const (
	JobVirgin      JobStatus = "Virgin"
	JobInProgress  JobStatus = "InProgress"
	JobDoneSuccess JobStatus = "DoneSuccess"
	JobDoneError   JobStatus = "DoneError"
	JobInterrupted JobStatus = "Interrupted"
)

type JobScenarioStatus string

const (
	JobScenarioVirgin      JobScenarioStatus = "Virgin"
	JobScenarioInProgress  JobScenarioStatus = "InProgress"
	JobScenarioDoneSuccess JobScenarioStatus = "DoneSuccess"
	JobScenarioDoneError   JobScenarioStatus = "DoneError"
	JobScenarioInterrupted JobScenarioStatus = "Interrupted"
)

type JobCreation struct {
	DelugeID string `json:"delugeId"`
	Webhook  string `json:"webhook"`
}

type Job struct {
	ID             string                  `json:"id"`
	DelugeID       string                  `json:"delugeId"`
	DelugeName     string                  `json:"delugeName"`
	Status         JobStatus               `json:"status"`
	GlobalDuration time.Duration           `json:"globalDuration"`
	Scenarios      map[string]*JobScenario `json:"scenarios"`
}

type JobLite struct {
	ID         string    `json:"id"`
	DelugeID   string    `json:"delugeId"`
	DelugeName string    `json:"delugeName"`
	Status     JobStatus `json:"status"`
}

type JobScenario struct {
	Name              string            `json:"name"`
	IterationDuration time.Duration     `json:"iterationDuration"`
	Status            JobScenarioStatus `json:"status"`
	Errors            []*object.Error   `json:"errors"`
	Report            reporting.Report  `json:"report"`
}

func mapDeluge(d *core.RunnableDeluge) *Job {
	d.Mutex.Lock()
	dDTO := &Job{
		DelugeID:       d.GetDelugeDefinition().ID,
		DelugeName:     d.GetDelugeDefinition().Name,
		GlobalDuration: d.GetGlobalDuration(),
		Status:         mapDelugeStatus(d.Status),
		Scenarios:      make(map[string]*JobScenario),
	}
	d.Mutex.Unlock()
	for scID, sc := range d.Scenarios {
		sc.Mutex.Lock()
		dDTO.Scenarios[scID] = mapScenario(sc)
		sc.Mutex.Unlock()
	}
	return dDTO
}

func mapDelugeLite(d *core.RunnableDeluge) *JobLite {
	d.Mutex.Lock()
	dDTO := &JobLite{
		DelugeID:   d.GetDelugeDefinition().ID,
		DelugeName: d.GetDelugeDefinition().Name,
		Status:     mapDelugeStatus(d.Status),
	}
	d.Mutex.Unlock()
	return dDTO
}

func mapScenario(sc *core.RunnableScenario) *JobScenario {
	httpReporter := &reporting.HTTPReporter{}
	return &JobScenario{
		Name:              sc.GetScenarioDefinition().Name,
		IterationDuration: sc.IterationDuration,
		Errors:            sc.Errors,
		Report:            httpReporter.Report(sc.Records),
		Status:            mapScenarioStatus(sc.Status),
	}
}

func mapScenarioStatus(st core.ScenarioStatus) JobScenarioStatus {
	switch st {
	case core.ScenarioVirgin:
		return JobScenarioVirgin
	case core.ScenarioInProgress:
		return JobScenarioInProgress
	case core.ScenarioDoneSuccess:
		return JobScenarioDoneSuccess
	case core.ScenarioDoneError:
		return JobScenarioDoneError
	case core.ScenarioInterrupted:
		return JobScenarioInterrupted
	}
	panic(errors.New(fmt.Sprintf("Invalid scenario status %d", st)))
}

func mapDelugeStatus(st core.DelugeStatus) JobStatus {
	switch st {
	case core.DelugeVirgin:
		return JobVirgin
	case core.DelugeInProgress:
		return JobInProgress
	case core.DelugeDoneSuccess:
		return JobDoneSuccess
	case core.DelugeDoneError:
		return JobDoneError
	case core.DelugeInterrupted:
		return JobInterrupted
	}
	panic(errors.New(fmt.Sprintf("Invalid deluge status %d", st)))
}
