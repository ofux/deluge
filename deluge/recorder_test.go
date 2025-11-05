package deluge

import (
	"sync"
	"testing"
	"time"
)

func Test__Verifier__Recorder(t *testing.T) {

	t.Run("Records 1 value", func(t *testing.T) {
		recorder := NewRecorder()

		recorder.Record(0, "foo", 1000)

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatal(err.Error())
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
		recorder := NewRecorder()

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
			t.Fatal(err.Error())
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
		recorder := NewRecorder()

		recorder.Record(42, "foo", 1000)

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatal(err.Error())
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
		recorder := NewRecorder()

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
			t.Fatal(err.Error())
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

func Test__Verifier__Recorder_Errors(t *testing.T) {

	t.Run("Get records on a running recorder", func(t *testing.T) {
		recorder := NewRecorder()

		recorder.Record(0, "foo", 1000)

		_, err := recorder.GetRecords()
		if err == nil {
			t.Error("Excpected non-nil error")
		}
	})
}

func Test__Verifier__Recorder_Values(t *testing.T) {

	t.Run("Records correct values not just counts", func(t *testing.T) {
		recorder := NewRecorder()

		// Record specific values that we can verify
		recorder.Record(0, "foo", 1000)
		recorder.Record(0, "foo", 2000)
		recorder.Record(0, "foo", 3000)

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatal(err.Error())
		}

		result, ok := results["foo"]
		if !ok {
			t.Fatalf("Expected to have some records for '%s'", "foo")
		}

		histogram := result[0]
		if histogram.TotalCount() != 3 {
			t.Errorf("Expected totalCount = 3, got %d", histogram.TotalCount())
		}

		// Check that actual values were recorded correctly
		// Min should be around 1000, max around 3000, mean around 2000
		// HDR histogram uses bucketing, so allow small rounding tolerance
		min := histogram.Min()
		if min < 999 || min > 1001 {
			t.Errorf("Expected min around 1000, got %d", min)
		}

		max := histogram.Max()
		if max < 2999 || max > 3001 {
			t.Errorf("Expected max around 3000, got %d", max)
		}

		mean := histogram.Mean()
		if mean < 1900 || mean > 2100 {
			t.Errorf("Expected mean around 2000, got %.2f", mean)
		}
	})

	t.Run("Records large values correctly", func(t *testing.T) {
		recorder := NewRecorder()

		// Record large values (> 1 second in microseconds) to test histogram max value
		largeValue := int64(2000000000) // 2000 seconds in microseconds
		recorder.Record(0, "foo", largeValue)

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatal(err.Error())
		}

		result, ok := results["foo"]
		if !ok {
			t.Fatalf("Expected to have some records for '%s'", "foo")
		}

		histogram := result[0]
		if histogram.TotalCount() != 1 {
			t.Errorf("Expected totalCount = 1, got %d", histogram.TotalCount())
		}

		// Verify the large value was recorded correctly
		max := histogram.Max()
		if max < largeValue-1000000 || max > largeValue+1000000 {
			t.Errorf("Expected max around %d, got %d", largeValue, max)
		}
	})

	t.Run("Records values with sufficient precision", func(t *testing.T) {
		recorder := NewRecorder()

		// Record values that are close together to test precision
		// With precision 3, we should be able to distinguish these
		recorder.Record(0, "foo", 1000)
		recorder.Record(0, "foo", 1001)
		recorder.Record(0, "foo", 1002)

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatal(err.Error())
		}

		result, ok := results["foo"]
		if !ok {
			t.Fatalf("Expected to have some records for '%s'", "foo")
		}

		histogram := result[0]

		// With proper precision, mean should be close to 1001
		mean := histogram.Mean()
		if mean < 1000 || mean > 1002 {
			t.Errorf("Expected mean around 1001 with good precision, got %.2f", mean)
		}

		// Check that we can distinguish between min and max
		min := histogram.Min()
		max := histogram.Max()
		if max-min < 1 {
			t.Errorf("Expected to distinguish between values with precision, got min=%d max=%d", min, max)
		}
	})
}

func Test__Verifier__Recorder_MultipleIDs(t *testing.T) {

	t.Run("Records values for multiple different IDs", func(t *testing.T) {
		recorder := NewRecorder()

		// Record values for different IDs to test map initialization
		recorder.Record(0, "foo", 1000)
		recorder.Record(0, "bar", 2000)
		recorder.Record(0, "baz", 3000)

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatal(err.Error())
		}

		// Verify each ID has its own histogram
		fooResult, ok := results["foo"]
		if !ok {
			t.Fatalf("Expected to have records for 'foo'")
		}
		if fooResult[0].TotalCount() != 1 {
			t.Errorf("Expected foo totalCount = 1, got %d", fooResult[0].TotalCount())
		}
		if fooResult[0].Mean() < 900 || fooResult[0].Mean() > 1100 {
			t.Errorf("Expected foo mean around 1000, got %.2f", fooResult[0].Mean())
		}

		barResult, ok := results["bar"]
		if !ok {
			t.Fatalf("Expected to have records for 'bar'")
		}
		if barResult[0].TotalCount() != 1 {
			t.Errorf("Expected bar totalCount = 1, got %d", barResult[0].TotalCount())
		}
		if barResult[0].Mean() < 1900 || barResult[0].Mean() > 2100 {
			t.Errorf("Expected bar mean around 2000, got %.2f", barResult[0].Mean())
		}

		bazResult, ok := results["baz"]
		if !ok {
			t.Fatalf("Expected to have records for 'baz'")
		}
		if bazResult[0].TotalCount() != 1 {
			t.Errorf("Expected baz totalCount = 1, got %d", bazResult[0].TotalCount())
		}
		if bazResult[0].Mean() < 2900 || bazResult[0].Mean() > 3100 {
			t.Errorf("Expected baz mean around 3000, got %.2f", bazResult[0].Mean())
		}
	})

	t.Run("Records to new IDs at different iterations without gaps", func(t *testing.T) {
		recorder := NewRecorder()

		// Test map initialization with IDs starting at iteration 0
		// This should ensure proper map entry creation
		recorder.Record(0, "id1", 100)
		recorder.Record(0, "id2", 200)
		recorder.Record(0, "id1", 101)
		recorder.Record(0, "id2", 201)

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatal(err.Error())
		}

		// Verify both IDs have correct count and all histograms exist
		if len(results) != 2 {
			t.Errorf("Expected 2 IDs in results, got %d", len(results))
		}

		id1Result, ok := results["id1"]
		if !ok {
			t.Fatalf("Expected to have records for 'id1'")
		}
		if len(id1Result) != 1 {
			t.Errorf("Expected id1 to have 1 iteration, got %d", len(id1Result))
		}
		if id1Result[0].TotalCount() != 2 {
			t.Errorf("Expected id1 totalCount = 2, got %d", id1Result[0].TotalCount())
		}

		id2Result, ok := results["id2"]
		if !ok {
			t.Fatalf("Expected to have records for 'id2'")
		}
		if len(id2Result) != 1 {
			t.Errorf("Expected id2 to have 1 iteration, got %d", len(id2Result))
		}
		if id2Result[0].TotalCount() != 2 {
			t.Errorf("Expected id2 totalCount = 2, got %d", id2Result[0].TotalCount())
		}
	})
}

func Test__Verifier__Recorder_CloseTransitions(t *testing.T) {

	t.Run("Close properly transitions to terminated state", func(t *testing.T) {
		recorder := NewRecorder()

		recorder.Record(0, "foo", 1000)

		// Close should properly transition through states and make GetRecords available
		recorder.Close()

		// GetRecords should work immediately after Close without error
		results, err := recorder.GetRecords()
		if err != nil {
			t.Errorf("Expected no error after Close, got: %v", err)
		}

		if results == nil {
			t.Error("Expected non-nil results after Close")
		}

		// Calling GetRecords again should still work (state should remain TERMINATED)
		results2, err2 := recorder.GetRecords()
		if err2 != nil {
			t.Errorf("Expected no error on second GetRecords call, got: %v", err2)
		}

		if results2 == nil {
			t.Error("Expected non-nil results on second GetRecords call")
		}
	})

	t.Run("Close is idempotent and safe to call multiple times", func(t *testing.T) {
		recorder := NewRecorder()

		recorder.Record(0, "foo", 1000)

		// First close
		recorder.Close()

		// Second close should not panic or cause issues
		recorder.Close()

		// Third close
		recorder.Close()

		// Should still be able to get records after multiple Close calls
		results, err := recorder.GetRecords()
		if err != nil {
			t.Errorf("Expected no error after multiple Close calls, got: %v", err)
		}

		if results == nil {
			t.Error("Expected non-nil results")
		}

		result, ok := results["foo"]
		if !ok || result[0].TotalCount() != 1 {
			t.Error("Expected results to be intact after multiple Close calls")
		}
	})
}
