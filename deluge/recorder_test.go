package deluge

import (
	"sync"
	"testing"
	"time"
)

func TestQueuedRecorder(t *testing.T) {

	t.Run("Records 1 value", func(t *testing.T) {
		recorder := NewRecorder(10)

		recorder.Record(0, "foo", 1000)

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		result, ok := results["foo"]
		if !ok {
			t.Fatalf("Expected to have some records for '%s'", "foo")
		}
		if len(result) != 1 {
			t.Fatalf("Expected to have %d records for '%s', got %d", 1, "foo", len(result))
		}
		if result[0].TotalCount() != 1 {
			t.Errorf("Expected to have totalCount = %d, got %d", 1, result[0].TotalCount())
		}
	})

	t.Run("Records 100 values simultaneously on the same iteration", func(t *testing.T) {
		const concurrent = 100
		recorder := NewRecorder(10)

		var waitg sync.WaitGroup
		for i := 0; i < concurrent; i++ {
			waitg.Add(1)
			go func(i int) {
				defer waitg.Done()
				recorder.Record(0, "foo", int64(100*i))
			}(i)
		}
		waitg.Wait()

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		result, ok := results["foo"]
		if !ok {
			t.Fatalf("Expected to have some records for '%s'", "foo")
		}
		if len(result) != 1 {
			t.Fatalf("Expected to have %d records for '%s', got %d", 1, "foo", len(result))
		}
		if result[0].TotalCount() != concurrent {
			t.Errorf("Expected to have totalCount = %d, got %d", concurrent, result[0].TotalCount())
		}
	})

	t.Run("Records 1 value at a given iteration", func(t *testing.T) {
		recorder := NewRecorder(10)

		recorder.Record(42, "foo", 1000)

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatalf(err.Error())
		}

		result, ok := results["foo"]
		if !ok {
			t.Fatalf("Expected to have some records for '%s'", "foo")
		}
		if len(result) != 43 {
			t.Fatalf("Expected to have %d records for '%s', got %d", 43, "foo", len(result))
		}
		if result[0].TotalCount() != 0 {
			t.Errorf("Expected to have totalCount = %d, got %d", 0, result[0].TotalCount())
		}
		if result[42].TotalCount() != 1 {
			t.Errorf("Expected to have totalCount = %d, got %d", 1, result[42].TotalCount())
		}
	})

	t.Run("Records 100 values simultaneously on multiple iterations", func(t *testing.T) {
		const concurrent = 100
		const iterCount = 100
		recorder := NewRecorder(10)

		var waitg sync.WaitGroup
		for i := 0; i < concurrent; i++ {
			waitg.Add(1)
			go func(i int) {
				defer waitg.Done()
				for j := 0; j < iterCount; j++ {
					recorder.Record(j, "foo", int64(100*i))
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

		result, ok := results["foo"]
		if !ok {
			t.Fatalf("Expected to have some records for '%s'", "foo")
		}
		if len(result) != iterCount {
			t.Fatalf("Expected to have %d records for '%s', got %d", iterCount, "foo", len(result))
		}
		for j := 0; j < iterCount; j++ {
			if result[j].TotalCount() != concurrent {
				t.Errorf("Expected to have totalCount = %d, got %d for iteration %d", concurrent, result[j].TotalCount(), j)
			}
		}
	})
}

func TestQueuedRecorderErrors(t *testing.T) {

	t.Run("Records a value on a closed recorder", func(t *testing.T) {
		recorder := NewRecorder(10)

		recorder.Close()

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected a panic")
			}
		}()
		// panic!!!
		recorder.Record(0, "foo", 1000)
	})

	t.Run("Get records on a running recorder", func(t *testing.T) {
		recorder := NewRecorder(10)

		recorder.Record(0, "foo", 1000)

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
