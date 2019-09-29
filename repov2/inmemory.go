package repov2

import (
	hdr "github.com/codahale/hdrhistogram"
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/core/status"
	"github.com/ofux/deluge/dsl/object"
	"sync"
	"time"
)

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

type Repository struct {
	delugeDefinitions map[string]*PersistedDeluge
	mutDeluges        *sync.Mutex

	scenarioDefinitions map[string]*PersistedScenario
	mutScenarios        *sync.Mutex

	jobShells    map[string]*PersistedJobShell
	mutJobShells *sync.Mutex

	workerReports    map[string]*PersistedWorkerReport
	mutWorkerReports *sync.Mutex
}

var Instance = NewInMemoryRepository()

func NewInMemoryRepository() *Repository {
	return &Repository{
		delugeDefinitions:   make(map[string]*PersistedDeluge),
		mutDeluges:          &sync.Mutex{},
		scenarioDefinitions: make(map[string]*PersistedScenario),
		mutScenarios:        &sync.Mutex{},
		jobShells:           make(map[string]*PersistedJobShell),
		mutJobShells:        &sync.Mutex{},
		workerReports:       make(map[string]*PersistedWorkerReport),
		mutWorkerReports:    &sync.Mutex{},
	}
}

func (r *Repository) SaveDeluge(deluge *PersistedDeluge) error {
	r.mutDeluges.Lock()
	defer r.mutDeluges.Unlock()
	r.delugeDefinitions[deluge.ID] = deluge
	return nil
}

func (r *Repository) GetDeluge(id string) (*PersistedDeluge, bool) {
	r.mutDeluges.Lock()
	defer r.mutDeluges.Unlock()
	def, ok := r.delugeDefinitions[id]
	return def, ok
}

func (r *Repository) GetAllDeluges() []*PersistedDeluge {
	r.mutDeluges.Lock()
	defer r.mutDeluges.Unlock()
	all := make([]*PersistedDeluge, 0, len(r.delugeDefinitions))
	for _, v := range r.delugeDefinitions {
		all = append(all, v)
	}
	return all
}

func (r *Repository) DeleteDeluge(id string) bool {
	r.mutDeluges.Lock()
	defer r.mutDeluges.Unlock()
	if _, ok := r.delugeDefinitions[id]; ok {
		delete(r.delugeDefinitions, id)
		return true
	}
	return false
}

// ======

func (r *Repository) SaveScenario(scenario *PersistedScenario) error {
	r.mutScenarios.Lock()
	defer r.mutScenarios.Unlock()
	r.scenarioDefinitions[scenario.ID] = scenario
	return nil
}

func (r *Repository) GetScenario(id string) (*PersistedScenario, bool) {
	r.mutScenarios.Lock()
	defer r.mutScenarios.Unlock()
	def, ok := r.scenarioDefinitions[id]
	return def, ok
}

func (r *Repository) GetDelugeScenarios(ids []string) map[string]*PersistedScenario {
	r.mutScenarios.Lock()
	defer r.mutScenarios.Unlock()
	delugeScenarios := make(map[string]*PersistedScenario)
	for _, id := range ids {
		if scenario, ok := r.scenarioDefinitions[id]; ok {
			delugeScenarios[id] = scenario
		}
	}
	return delugeScenarios
}

func (r *Repository) GetAllScenarios() []*PersistedScenario {
	r.mutScenarios.Lock()
	defer r.mutScenarios.Unlock()
	all := make([]*PersistedScenario, 0, len(r.scenarioDefinitions))
	for _, v := range r.scenarioDefinitions {
		all = append(all, v)
	}
	return all
}

func (r *Repository) DeleteScenario(id string) bool {
	r.mutScenarios.Lock()
	defer r.mutScenarios.Unlock()
	if _, ok := r.scenarioDefinitions[id]; ok {
		delete(r.scenarioDefinitions, id)
		return true
	}
	return false
}

// =======

func (r *Repository) SaveJobShell(jobShell *PersistedJobShell) error {
	r.mutJobShells.Lock()
	defer r.mutJobShells.Unlock()
	r.jobShells[jobShell.ID] = jobShell
	return nil
}

func (r *Repository) GetJobShell(id string) (*PersistedJobShell, bool) {
	r.mutJobShells.Lock()
	defer r.mutJobShells.Unlock()
	jobShell, ok := r.jobShells[id]
	return jobShell, ok
}

func (r *Repository) GetAllJobShell() []*PersistedJobShell {
	r.mutJobShells.Lock()
	defer r.mutJobShells.Unlock()
	all := make([]*PersistedJobShell, 0, len(r.jobShells))
	for _, v := range r.jobShells {
		all = append(all, v)
	}
	return all
}

// WorkerReports

func (r *Repository) SaveWorkerReport(workerReport *PersistedWorkerReport) error {
	r.mutWorkerReports.Lock()
	defer r.mutWorkerReports.Unlock()
	r.workerReports[workerReport.WorkerID] = workerReport
	return nil
}

func (r *Repository) GetWorkerReports(jobID string) []*PersistedWorkerReport {
	r.mutWorkerReports.Lock()
	defer r.mutWorkerReports.Unlock()
	var reports []*PersistedWorkerReport
	for _, v := range r.workerReports {
		if v.JobID == jobID {
			reports = append(reports, v)
		}
	}
	return reports
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
