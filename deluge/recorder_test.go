package deluge

import (
	"sync"
	"testing"
	"time"

	hdr "github.com/ofux/hdrhistogram"
)

func Test__Verifier__Recorder(t *testing.T) {

	t.Run("Records 1 value", func(t *testing.T) {
		recorder := NewRecorder(DefaultHistogramConstructor())

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
		recorder := NewRecorder(DefaultHistogramConstructor())

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
		recorder := NewRecorder(DefaultHistogramConstructor())

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
		recorder := NewRecorder(DefaultHistogramConstructor())

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

	t.Run("Records correct value - verifies actual value not just count", func(t *testing.T) {
		recorder := NewRecorder(DefaultHistogramConstructor())

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

		// Type assert to access value methods
		h, ok := result[0].(*hdr.Histogram)
		if !ok {
			t.Fatalf("Expected histogram to be *hdr.Histogram")
		}

		if h.Mean() != 1000 {
			t.Errorf("Expected mean = %d, got %f", 1000, h.Mean())
		}
		if h.Min() != 1000 {
			t.Errorf("Expected min = %d, got %d", 1000, h.Min())
		}
		if h.Max() != 1000 {
			t.Errorf("Expected max = %d, got %d", 1000, h.Max())
		}
	})

	t.Run("Records multiple different values - verifies correct statistics", func(t *testing.T) {
		recorder := NewRecorder(DefaultHistogramConstructor())

		// Record values: 100, 200, 300
		recorder.Record(0, "foo", 100)
		recorder.Record(0, "foo", 200)
		recorder.Record(0, "foo", 300)

		recorder.Close()

		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatal(err.Error())
		}

		result, ok := results["foo"]
		if !ok {
			t.Fatalf("Expected to have some records for '%s'", "foo")
		}

		h, ok := result[0].(*hdr.Histogram)
		if !ok {
			t.Fatalf("Expected histogram to be *hdr.Histogram")
		}

		if h.TotalCount() != 3 {
			t.Errorf("Expected totalCount = %d, got %d", 3, h.TotalCount())
		}
		if h.Min() != 100 {
			t.Errorf("Expected min = %d, got %d", 100, h.Min())
		}
		if h.Max() != 300 {
			t.Errorf("Expected max = %d, got %d", 300, h.Max())
		}
		expectedMean := 200.0
		if h.Mean() != expectedMean {
			t.Errorf("Expected mean = %f, got %f", expectedMean, h.Mean())
		}
	})

	t.Run("Records values concurrently - verifies actual values recorded", func(t *testing.T) {
		const concurrent = 10
		recorder := NewRecorder(DefaultHistogramConstructor())

		var waitg sync.WaitGroup
		for i := 0; i < concurrent; i++ {
			waitg.Add(1)
			go func(i int) {
				defer waitg.Done()
				// Each goroutine records a unique value: 1000, 2000, 3000, ...
				recorder.Record(0, "foo", int64(1000*(i+1)))
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

		h, ok := result[0].(*hdr.Histogram)
		if !ok {
			t.Fatalf("Expected histogram to be *hdr.Histogram")
		}

		if h.TotalCount() != concurrent {
			t.Errorf("Expected totalCount = %d, got %d", concurrent, h.TotalCount())
		}
		// Min should be 1000
		if h.Min() != 1000 {
			t.Errorf("Expected min = %d, got %d", 1000, h.Min())
		}
		// Max should be around 10000 (HDR histogram has slight precision loss)
		if h.Max() < 10000 || h.Max() > 10100 {
			t.Errorf("Expected max around 10000, got %d", h.Max())
		}
		// Mean should be around 5500 (with some tolerance for HDR precision)
		expectedMean := 5500.0
		tolerance := 100.0
		if h.Mean() < expectedMean-tolerance || h.Mean() > expectedMean+tolerance {
			t.Errorf("Expected mean around %f (±%f), got %f", expectedMean, tolerance, h.Mean())
		}
		// If values were recorded as 0, mean would be 0
		if h.Mean() < 100 {
			t.Errorf("Mean is too low (%f), values may not have been recorded correctly", h.Mean())
		}
	})
}
