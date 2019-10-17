package api

import (
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/core/reporting"
	"github.com/ofux/deluge/core/status"
	"github.com/ofux/deluge/dsl/object"
	"github.com/ofux/deluge/repov2"
	"github.com/pkg/errors"
	"time"
)

type JobCreation struct {
	DelugeID string `json:"delugeId"`
	Webhook  string `json:"webhook"`
}

type JobMetadata struct {
	ID       string `json:"id"`
	DelugeID string `json:"delugeId"`
	Webhook  string `json:"webhook"`
}

type Job struct {
	ID             string                  `json:"id"`
	DelugeID       string                  `json:"delugeId"`
	DelugeName     string                  `json:"delugeName"`
	Status         status.DelugeStatus     `json:"status"`
	GlobalDuration time.Duration           `json:"globalDuration"`
	Scenarios      map[string]*JobScenario `json:"scenarios"`
}

type JobScenario struct {
	ID                string                `json:"scenarioId"`
	Name              string                `json:"name"`
	IterationDuration time.Duration         `json:"iterationDuration"`
	Status            status.ScenarioStatus `json:"status"`
	Errors            []*object.Error       `json:"errors"`
	Report            reporting.Report      `json:"report"`
}

func mapDeluge(job *repov2.PersistedJobShell, deluge *repov2.PersistedDeluge, scenarioDefs map[string]*repov2.PersistedScenario, workerReports []*repov2.PersistedWorkerReport) (*Job, error) {

	dDTO := &Job{
		ID:         job.ID,
		DelugeID:   job.DelugeID,
		DelugeName: "not found",
	}

	if deluge != nil {
		dDTO.DelugeName = deluge.Name
		dDTO.GlobalDuration = deluge.GlobalDuration
	}

	delugeStatus := status.DelugeVirgin
	scenariosStatus := make(map[string]status.ScenarioStatus)
	scenariosErrors := make(map[string][]*object.Error)
	scenariosIterationDurations := make(map[string]time.Duration)
	scenariosRecords := make(map[string]*recording.HTTPRecordsOverTime)
	httpReporter := &reporting.HTTPReporter{}

	// Merge records
	for _, wr := range workerReports {
		delugeStatus = status.MergeDelugeStatuses(delugeStatus, wr.Status)
		for scenarioID, scenario := range wr.Scenarios {
			scenariosStatus[scenarioID] = status.MergeScenarioStatuses(scenariosStatus[scenarioID], scenario.Status)
			scenariosErrors[scenarioID] = append(scenariosErrors[scenarioID], scenario.Errors...)
			scenariosIterationDurations[scenarioID] = scenario.IterationDuration
			rec, err := recording.MapPersistedHTTPRecords(scenario.Records)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to map scenario %s of worker %s of job %s", scenarioID, wr.WorkerID, wr.JobID)
			}
			scenariosRecords[scenarioID] = recording.MergeHTTPRecordsOverTime(scenariosRecords[scenarioID], rec)
		}
	}

	jobScenarios := make(map[string]*JobScenario)
	for scenarioID, scenarioStatus := range scenariosStatus {
		jobScenario := &JobScenario{
			IterationDuration: scenariosIterationDurations[scenarioID],
			Status:            scenarioStatus,
			Errors:            scenariosErrors[scenarioID],
			Report:            httpReporter.Report(scenariosRecords[scenarioID]),
		}
		if scenarioDefs != nil {
			if scenarioDef, ok := scenarioDefs[scenarioID]; ok {
				jobScenario.Name = scenarioDef.Name
			}
		}
		jobScenarios[scenarioID] = jobScenario
	}

	dDTO.Status = delugeStatus
	dDTO.Scenarios = jobScenarios
	return dDTO, nil
}
