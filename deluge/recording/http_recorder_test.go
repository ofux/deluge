package recording

import (
	"github.com/ofux/deluge/deluge/recording/recordingtest"
	"sync"
	"testing"
	"time"
)

func TestHTTPRecorder(t *testing.T) {

	t.Run("Records 1 Value", func(t *testing.T) {
		recorder := NewHTTPRecorder(10)

		recorder.Record(&HTTPRecordEntry{
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

		result := results.PerIteration[0]
		recordingtest.CheckHTTPRecord(t, result, "foo", 1, 200, Ok)
	})

	t.Run("Records 100 values simultaneously on the same Iteration", func(t *testing.T) {
		const concurrent = 100
		recorder := NewHTTPRecorder(10)

		var waitg sync.WaitGroup
		for i := 0; i < concurrent; i++ {
			waitg.Add(1)
			go func(i int) {
				defer waitg.Done()
				recorder.Record(&HTTPRecordEntry{
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

		result := results.PerIteration[0]
		recordingtest.CheckHTTPRecord(t, result, "foo", concurrent, 200, Ok)
	})

	t.Run("Records 1 Value at a given Iteration", func(t *testing.T) {
		recorder := NewHTTPRecorder(10)

		recorder.Record(&HTTPRecordEntry{
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

		result := results.PerIteration[42]
		if len(results.PerIteration) != 43 {
			t.Fatalf("Expected to have %d iterations, got %d", 43, len(results.PerIteration))
		}
		recordingtest.CheckHTTPRecord(t, result, "foo", 1, 200, Ok)
	})

	t.Run("Records 100 values simultaneously on multiple iterations", func(t *testing.T) {
		const concurrent = 100
		const iterCount = 100
		recorder := NewHTTPRecorder(10)

		var waitg sync.WaitGroup
		for i := 0; i < concurrent; i++ {
			waitg.Add(1)
			go func(i int) {
				defer waitg.Done()
				for j := 0; j < iterCount; j++ {
					recorder.Record(&HTTPRecordEntry{
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

		recordingtest.CheckHTTPRecord(t, results.Global, "foo", 1, 200, Ok)
		for j := 0; j < iterCount; j++ {
			recordingtest.CheckHTTPRecord(t, results.PerIteration[j], "foo", 1, 200, Ok)
		}
	})
}

func TestHTTPRecorderErrors(t *testing.T) {

	t.Run("Get records on a running httpRecorder", func(t *testing.T) {
		recorder := NewHTTPRecorder(10)

		recorder.Record(&HTTPRecordEntry{
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
