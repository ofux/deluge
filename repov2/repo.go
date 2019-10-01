package repov2

import (
	hdr "github.com/codahale/hdrhistogram"
	"github.com/ofux/deluge/core/recording"
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

type PersistedWorkerScenarioReport struct {
	Status            status.ScenarioStatus
	Errors            []*object.Error
	IterationDuration time.Duration
	Records           *PersistedHTTPRecordsOverTime
}

type PersistedHTTPRecordsOverTime struct {
	Global       *PersistedHTTPRecord
	PerIteration []*PersistedHTTPRecord
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

func MapHTTPRecords(records *recording.HTTPRecordsOverTime) (*PersistedHTTPRecordsOverTime, error) {
	report := &PersistedHTTPRecordsOverTime{
		Global:       mapHTTPRecord(records.Global),
		PerIteration: make([]*PersistedHTTPRecord, 0, 16),
	}
	for _, v := range records.PerIteration {
		report.PerIteration = append(report.PerIteration, mapHTTPRecord(v))
	}

	return report, nil
}

func mapHTTPRecord(rec *recording.HTTPRecord) *PersistedHTTPRecord {
	st := &PersistedHTTPRecord{
		PersistedHTTPRequestRecord: *mapHTTPRequestRecord(&(rec.HTTPRequestRecord)),
		PerRequests:                make(map[string]*PersistedHTTPRequestRecord),
	}
	for k, v := range rec.PerRequests {
		st.PerRequests[k] = mapHTTPRequestRecord(v)
	}
	return st
}

func mapHTTPRequestRecord(rec *recording.HTTPRequestRecord) *PersistedHTTPRequestRecord {
	st := &PersistedHTTPRequestRecord{
		Global:    rec.Global.Export(),
		PerStatus: make(map[int]*hdr.Snapshot),
		PerOkKo:   make(map[OkKo]*hdr.Snapshot),
	}
	for k, v := range rec.PerStatus {
		st.PerStatus[k] = v.Export()
	}
	for k, v := range rec.PerOkKo {
		key := Ok
		if k == recording.Ko {
			key = Ko
		}
		st.PerOkKo[key] = v.Export()
	}
	return st
}

func MapPersistedHTTPRecords(records *PersistedHTTPRecordsOverTime) *recording.HTTPRecordsOverTime {
	if records == nil {
		return nil
	}
	report := &recording.HTTPRecordsOverTime{
		Global:       mapPersistedHTTPRecord(records.Global),
		PerIteration: make([]*recording.HTTPRecord, 0, 16),
	}
	for _, v := range records.PerIteration {
		report.PerIteration = append(report.PerIteration, mapPersistedHTTPRecord(v))
	}

	return report
}

func mapPersistedHTTPRecord(rec *PersistedHTTPRecord) *recording.HTTPRecord {
	st := &recording.HTTPRecord{
		HTTPRequestRecord: *mapPersistedHTTPRequestRecord(&(rec.PersistedHTTPRequestRecord)),
		PerRequests:       make(map[string]*recording.HTTPRequestRecord),
	}
	for k, v := range rec.PerRequests {
		st.PerRequests[k] = mapPersistedHTTPRequestRecord(v)
	}
	return st
}

func mapPersistedHTTPRequestRecord(rec *PersistedHTTPRequestRecord) *recording.HTTPRequestRecord {
	st := &recording.HTTPRequestRecord{
		Global:    hdr.Import(rec.Global),
		PerStatus: make(map[int]*hdr.Histogram),
		PerOkKo:   make(map[recording.OkKo]*hdr.Histogram),
	}
	for k, v := range rec.PerStatus {
		st.PerStatus[k] = hdr.Import(v)
	}
	for k, v := range rec.PerOkKo {
		key := recording.Ok
		if k == Ko {
			key = recording.Ko
		}
		st.PerOkKo[key] = hdr.Import(v)
	}
	return st
}
