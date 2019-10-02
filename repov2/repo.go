package repov2

import (
	hdr "github.com/codahale/hdrhistogram"
	"github.com/ofux/deluge/core/status"
	"github.com/ofux/deluge/dsl/object"
	"time"
)

var Instance Repository = NewInMemoryRepository()

type Repository interface {
	SaveDeluge(deluge *PersistedDeluge) error
	GetDeluge(id string) (*PersistedDeluge, bool)
	GetAllDeluges() []*PersistedDeluge
	DeleteDeluge(id string) bool

	SaveScenario(scenario *PersistedScenario) error
	GetScenario(id string) (*PersistedScenario, bool)
	GetDelugeScenarios(ids []string) map[string]*PersistedScenario
	GetAllScenarios() []*PersistedScenario
	DeleteScenario(id string) bool

	SaveJobShell(jobShell *PersistedJobShell) error
	GetJobShell(id string) (*PersistedJobShell, bool)
	GetAllJobShell() []*PersistedJobShell

	SaveWorkerReport(workerReport *PersistedWorkerReport) error
	GetJobWorkerReports(jobID string) []*PersistedWorkerReport
}

type PersistedDeluge struct {
	ID             string
	Name           string
	Script         string
	GlobalDuration time.Duration
	ScenarioIDs    []string
}

type PersistedScenario struct {
	ID     string
	Name   string
	Script string
}

type PersistedJobShell struct {
	ID       string
	DelugeID string
	Webhook  string
}

type PersistedWorkerReport struct {
	WorkerID  string
	JobID     string
	Status    status.DelugeStatus
	Scenarios map[string]*PersistedWorkerScenarioReport
}

func (wr *PersistedWorkerReport) GetID() string {
	return wr.WorkerID + "_" + wr.JobID
}

type PersistedWorkerScenarioReport struct {
	Status            status.ScenarioStatus
	Errors            []*object.Error
	IterationDuration time.Duration
	Records           *PersistedHTTPRecordsOverTime
}

type PersistedHTTPRecordsOverTime struct {
	Global   *PersistedHTTPRecord
	OverTime []*PersistedHTTPRecord
}

type PersistedHTTPRecord struct {
	PersistedHTTPRequestRecord
	PerRequests map[string]*PersistedHTTPRequestRecord
}

type PersistedHTTPRequestRecord struct {
	Global    *hdr.Snapshot
	PerStatus map[int]*hdr.Snapshot
	PerOkKo   map[OkKo]*hdr.Snapshot
}

type OkKo string

const (
	Ok OkKo = "Ok"
	Ko OkKo = "Ko"
)
