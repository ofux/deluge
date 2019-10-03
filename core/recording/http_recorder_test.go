package recording_test

import (
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/core/recording/recordingtest"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestHTTPRecorder(t *testing.T) {

	t.Run("Records 1 Value", func(t *testing.T) {
		recorder := recording.NewHTTPRecorder(1, 1)

		recorder.Record(&recording.HTTPRecordEntry{
			Iteration:  0,
			Name:       "foo",
			Value:      1000,
			StatusCode: 200,
		})

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		result := results.OverTime[0]
		recordingtest.CheckHTTPRecord(t, result, "foo", 1, 200, recording.Ok)
	})

	t.Run("Records 1 Value code 500", func(t *testing.T) {
		recorder := recording.NewHTTPRecorder(1, 1)

		recorder.Record(&recording.HTTPRecordEntry{
			Iteration:  0,
			Name:       "foo",
			Value:      1000,
			StatusCode: 500,
		})

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		result := results.OverTime[0]
		recordingtest.CheckHTTPRecord(t, result, "foo", 1, 500, recording.Ko)
	})

	t.Run("Records 100 values simultaneously on the same Iteration", func(t *testing.T) {
		const concurrent = 100
		recorder := recording.NewHTTPRecorder(1, concurrent)

		var waitg sync.WaitGroup
		for i := 0; i < concurrent; i++ {
			waitg.Add(1)
			go func(i int) {
				defer waitg.Done()
				recorder.Record(&recording.HTTPRecordEntry{
					Iteration:  0,
					Name:       "foo",
					Value:      int64(100 * i),
					StatusCode: 200,
				})
			}(i)
		}
		waitg.Wait()

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		result := results.OverTime[0]
		recordingtest.CheckHTTPRecord(t, result, "foo", concurrent, 200, recording.Ok)
	})

	t.Run("Records 1 Value at a given Iteration", func(t *testing.T) {
		recorder := recording.NewHTTPRecorder(1, 1)

		recorder.Record(&recording.HTTPRecordEntry{
			Iteration:  42,
			Name:       "foo",
			Value:      1000,
			StatusCode: 200,
		})

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		result := results.OverTime[42]
		if len(results.OverTime) != 43 {
			t.Fatalf("Expected to have %d iterations, got %d", 43, len(results.OverTime))
		}
		recordingtest.CheckHTTPRecord(t, result, "foo", 1, 200, recording.Ok)
	})

	t.Run("Records 100 values simultaneously on multiple iterations", func(t *testing.T) {
		const concurrent = 100
		const iterCount = 100
		recorder := recording.NewHTTPRecorder(iterCount, concurrent)

		var waitg sync.WaitGroup
		for i := 0; i < concurrent; i++ {
			waitg.Add(1)
			go func(i int) {
				defer waitg.Done()
				for j := 0; j < iterCount; j++ {
					recorder.Record(&recording.HTTPRecordEntry{
						Iteration:  j,
						Name:       "foo",
						Value:      int64(100 * i),
						StatusCode: 200,
					})
					time.Sleep(time.Millisecond) // just to simulate some "real" scenario
				}
			}(i)
		}
		waitg.Wait()

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		recordingtest.CheckHTTPRecord(t, results.Global, "foo", iterCount*concurrent, 200, recording.Ok)
		require.Len(t, results.OverTime, iterCount)
		for j := 0; j < iterCount; j++ {
			recordingtest.CheckHTTPRecord(t, results.OverTime[j], "foo", concurrent, 200, recording.Ok)
		}
	})

	t.Run("Records 100 values simultaneously and get snapshots at the same time", func(t *testing.T) {
		const concurrent = 50
		const iterCount = 100
		const iterCountSnapshotReaders = 5
		var sleepDurationPerIteration = 10 * time.Millisecond
		var sleepDurationPerIterationForSnapshotReaders = sleepDurationPerIteration * iterCount / iterCountSnapshotReaders
		recorder := recording.NewHTTPRecorder(iterCount, concurrent)

		var waitg sync.WaitGroup
		for i := 0; i < concurrent; i++ {
			waitg.Add(1)
			go func(i int) {
				defer waitg.Done()
				for j := 0; j < iterCount; j++ {
					recorder.Record(&recording.HTTPRecordEntry{
						Iteration:  j,
						Name:       "foo",
						Value:      int64(100 * i),
						StatusCode: 200,
					})
					time.Sleep(sleepDurationPerIteration) // just to simulate some "real" scenario
				}
			}(i)
		}

		for j := 0; j < iterCountSnapshotReaders; j++ {
			snapshotChan, err := recorder.GetRecordsSnapshot()
			require.NoError(t, err)
			snapshot := <-snapshotChan
			require.NoError(t, snapshot.Err)
			if len(snapshot.HTTPRecordsOverTime.OverTime) < j {
				t.Errorf("There should be more iterations in the latest snapshot")
			}
			time.Sleep(sleepDurationPerIterationForSnapshotReaders)
		}

		waitg.Wait()

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		recordingtest.CheckHTTPRecord(t, results.Global, "foo", iterCount*concurrent, 200, recording.Ok)
		for j := 0; j < iterCount; j++ {
			recordingtest.CheckHTTPRecord(t, results.OverTime[j], "foo", concurrent, 200, recording.Ok)
		}
	})

	t.Run("Records 100 values simultaneously on more than MaxOverTimeCount iterations", func(t *testing.T) {
		const concurrent = 10
		const iterCount = recording.MaxOverTimeCount * 3
		recorder := recording.NewHTTPRecorder(iterCount, concurrent)

		var waitg sync.WaitGroup
		for i := 0; i < concurrent; i++ {
			waitg.Add(1)
			go func(i int) {
				defer waitg.Done()
				for j := 0; j < iterCount; j++ {
					recorder.Record(&recording.HTTPRecordEntry{
						Iteration:  j,
						Name:       "foo",
						Value:      int64(100 * i),
						StatusCode: 200,
					})
					time.Sleep(time.Millisecond) // just to simulate some "real" scenario
				}
			}(i)
		}
		waitg.Wait()

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		recordingtest.CheckHTTPRecord(t, results.Global, "foo", int64(iterCount*concurrent), 200, recording.Ok)
		require.Len(t, results.OverTime, recording.MaxOverTimeCount)
		for j := 0; j < recording.MaxOverTimeCount; j++ {
			recordingtest.CheckHTTPRecord(t, results.OverTime[j], "foo", concurrent*3, 200, recording.Ok)
		}
	})
}

func TestHTTPRecorderErrors(t *testing.T) {

	t.Run("Get records on a running httpRecorder", func(t *testing.T) {
		recorder := recording.NewHTTPRecorder(1, 1)

		recorder.Record(&recording.HTTPRecordEntry{
			Iteration:  0,
			Name:       "foo",
			Value:      1000,
			StatusCode: 200,
		})

		_, err := recorder.GetRecords()
		if err == nil {
			t.Error("Excpected non-nil error")
		}
		const expectedError = "Cannot get records while recording. Did you forget to call the 'Close()' method?"
		if err.Error() != expectedError {
			t.Errorf("Excpected error message to be '%s', got '%s'", expectedError, err.Error())
		}
	})
}
