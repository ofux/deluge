package worker

import (
	"errors"
	"github.com/ofux/deluge/core"
	"github.com/ofux/deluge/core/status"
	"github.com/ofux/deluge/repov2"
	"github.com/ofux/docilemonkey/docilemonkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
	"time"
)

func Test_worker_start(t *testing.T) {
	t.Run("Run worker without error", func(t *testing.T) {
		rep := &repoMock{
			InMemoryRepository: *repov2.NewInMemoryRepository(),
		}

		srv := docilemonkey.NewTestServer()
		defer srv.Close()

		const reqName = "My request"
		saveScenario(t, `
		scenario("scenario-id", "My scenario", function (args, session) {

			assert(args["foo"] == "bar");

			if (session["count"] == null) {
				session["count"] = 0;
			}
			session["count"]++;
			assert(session["count"] > 0);

			http("`+reqName+`", {
				"url": "`+srv.URL+`/hello/toto?s=201",
				"method": "POST"
			});

		});`)

		saveDeluge(t, `
		deluge("deluge-id", "Some name", "2s", {
			"scenario-id": {
				"concurrent": 5,
				"delay": "400ms",
				"args": {
					"foo": "bar"
				}
			}
		});`)

		worker := newWorker("worker-id", &JobShell{
			ID:       "job-id",
			DelugeID: "deluge-id",
		}, rep)
		require.NotNil(t, worker)

		rep.SaveWorkerReportImpl = func(workerReport *repov2.PersistedWorkerReport) error {
			rep.IncSaveWorkerReportCall()
			assert.Equal(t, "job-id", workerReport.JobID)
			assert.Equal(t, "worker-id", workerReport.WorkerID)
			switch rep.GetSaveWorkerReportCall() {
			case 1:
				assert.Equal(t, status.DelugeVirgin, workerReport.Status)
			case 2:
				assert.Equal(t, status.DelugeInProgress, workerReport.Status)
			case 3:
				assert.Equal(t, status.DelugeDoneSuccess, workerReport.Status)
				assert.NotNil(t, workerReport.Scenarios["scenario-id"])
				sc := workerReport.Scenarios["scenario-id"]
				assert.Equal(t, status.ScenarioDoneSuccess, sc.Status)
				assert.Len(t, sc.Errors, 0)
			}
			return rep.InMemoryRepository.SaveWorkerReport(workerReport)
		}

		err := worker.start()
		require.NoError(t, err)

		// Wait for the deluge
		time.Sleep(2 * time.Second)
		for wait := 0 * time.Millisecond; wait < 5*time.Second && rep.GetSaveWorkerReportCall() < 3; wait += 100 * time.Millisecond {
			time.Sleep(100 * time.Millisecond)
		}
		assert.Equal(t, int32(3), rep.GetSaveWorkerReportCall())
	})

	t.Run("Run worker with some repository failures and retry", func(t *testing.T) {
		rep := &repoMock{
			InMemoryRepository: *repov2.NewInMemoryRepository(),
		}

		srv := docilemonkey.NewTestServer()
		defer srv.Close()

		const reqName = "My request"
		saveScenario(t, `
		scenario("scenario-id", "My scenario", function () {

			http("`+reqName+`", {
				"url": "`+srv.URL+`/hello/toto?s=201",
				"method": "POST"
			});

		});`)

		saveDeluge(t, `
		deluge("deluge-id", "Some name", "2s", {
			"scenario-id": {
				"concurrent": 5,
				"delay": "400ms",
				"args": {
					"foo": "bar"
				}
			}
		});`)

		worker := newWorker("worker-id", &JobShell{
			ID:       "job-id",
			DelugeID: "deluge-id",
		}, rep)
		require.NotNil(t, worker)

		worker.finalReportRetryCount = 1
		worker.finalReportRetryDelay = 10 * time.Millisecond

		rep.SaveWorkerReportImpl = func(workerReport *repov2.PersistedWorkerReport) error {
			rep.IncSaveWorkerReportCall()
			assert.Equal(t, "job-id", workerReport.JobID)
			assert.Equal(t, "worker-id", workerReport.WorkerID)
			switch rep.GetSaveWorkerReportCall() {
			case 1:
				assert.Equal(t, status.DelugeVirgin, workerReport.Status)
			case 2:
				assert.Equal(t, status.DelugeInProgress, workerReport.Status)
			case 3:
				assert.Equal(t, status.DelugeDoneSuccess, workerReport.Status)
				assert.NotNil(t, workerReport.Scenarios["scenario-id"])
				sc := workerReport.Scenarios["scenario-id"]
				assert.Equal(t, status.ScenarioDoneSuccess, sc.Status)
				assert.Len(t, sc.Errors, 0)
				return errors.New("oops! failed to push report to repo")
			case 4:
				assert.Equal(t, status.DelugeDoneSuccess, workerReport.Status)
				assert.NotNil(t, workerReport.Scenarios["scenario-id"])
				sc := workerReport.Scenarios["scenario-id"]
				assert.Equal(t, status.ScenarioDoneSuccess, sc.Status)
				assert.Len(t, sc.Errors, 0)
			}
			return rep.InMemoryRepository.SaveWorkerReport(workerReport)
		}

		err := worker.start()
		require.NoError(t, err)

		// Wait for the deluge
		time.Sleep(2 * time.Second)
		for wait := 0 * time.Millisecond; wait < 5*time.Second && rep.GetSaveWorkerReportCall() < 4; wait += 100 * time.Millisecond {
			time.Sleep(100 * time.Millisecond)
		}
		assert.Equal(t, int32(4), rep.GetSaveWorkerReportCall())
	})
}

func saveDeluge(t testing.TB, script string) *core.CompiledDeluge {
	t.Helper()
	compiled, err := core.CompileDeluge(script)
	if err != nil {
		t.Fatal(err)
	}
	err = repov2.Instance.SaveDeluge(compiled.MapToPersistedDeluge())
	if err != nil {
		t.Fatal(err)
	}
	return compiled
}

func saveScenario(t testing.TB, script string) *core.CompiledScenario {
	t.Helper()
	compiled, err := core.CompileScenario(script)
	if err != nil {
		t.Fatal(err)
	}
	err = repov2.Instance.SaveScenario((*repov2.PersistedScenario)(compiled.GetScenarioDefinition()))
	if err != nil {
		t.Fatal(err)
	}
	return compiled
}

type repoMock struct {
	repov2.InMemoryRepository

	SaveWorkerReportImpl      func(workerReport *repov2.PersistedWorkerReport) error
	saveWorkerReportCallCount int32

	GetJobWorkerReportsImpl func(jobID string) []*repov2.PersistedWorkerReport
}

func (r *repoMock) IncSaveWorkerReportCall() {
	atomic.AddInt32(&r.saveWorkerReportCallCount, 1)
}

func (r *repoMock) GetSaveWorkerReportCall() int32 {
	return atomic.LoadInt32(&r.saveWorkerReportCallCount)
}

func (r *repoMock) SaveWorkerReport(workerReport *repov2.PersistedWorkerReport) error {
	if r.SaveWorkerReportImpl == nil {
		return r.InMemoryRepository.SaveWorkerReport(workerReport)
	}
	return r.SaveWorkerReportImpl(workerReport)
}

func (r *repoMock) GetJobWorkerReports(jobID string) []*repov2.PersistedWorkerReport {
	if r.GetJobWorkerReportsImpl == nil {
		return r.InMemoryRepository.GetJobWorkerReports(jobID)
	}
	return r.GetJobWorkerReportsImpl(jobID)
}
