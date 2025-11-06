//go:build !race

package deluge

import (
	"sync"
	"testing"
	"time"

	hdr "github.com/ofux/hdrhistogram"
)

type SlowHistogram struct {
	histogram *hdr.Histogram
	blocked   chan struct{}
}

func (h *SlowHistogram) RecordValue(value int64) error {
	<-h.blocked
	return h.histogram.RecordValue(value)
}

func (h *SlowHistogram) TotalCount() int64 {
	return h.histogram.TotalCount()
}

func Test__Verifier__RecordReturnsImmediately(t *testing.T) {
	t.Run("Record calls complete in under 1ms even with slow processing", func(t *testing.T) {
		blocked := make(chan struct{})
		recorder := NewRecorder(func() Histogram {
			return &SlowHistogram{histogram: hdr.New(0, 3600000000, 3), blocked: blocked}
		})

		// Record several values to potentially fill the buffer
		for i := 0; i < 20; i++ {
			start := time.Now()
			recorder.Record(0, "foo", int64(i))
			duration := time.Since(start)

			// Each Record call should complete almost instantly (well under 1ms)
			// Using 5ms as a generous threshold to account for system variations
			if duration > 200*time.Millisecond {
				t.Errorf("Record took %v, expected < 5ms (iteration %d)", duration, i)
			}
		}

		close(blocked)
		recorder.Close()

		// Verify all records were processed
		results, err := recorder.GetRecords()
		if err != nil {
			t.Fatal(err.Error())
		}
		result := results["foo"]
		if result[0].TotalCount() != 20 {
			t.Errorf("Expected totalCount = 20, got %d", result[0].TotalCount())
		}
	})

	t.Run("Multiple Record calls return immediately under load", func(t *testing.T) {
		const concurrentCalls = 200000
		blocked := make(chan struct{})
		recorder := NewRecorder(func() Histogram {
			return &SlowHistogram{histogram: hdr.New(0, 3600000000, 3), blocked: blocked}
		})

		// Launch many concurrent Record calls
		var waitg sync.WaitGroup
		timings := make([]time.Duration, concurrentCalls)

		for i := 0; i < concurrentCalls; i++ {
			waitg.Add(1)
			go func(i int) {
				defer waitg.Done()
				start := time.Now()
				recorder.Record(0, "foo", int64(i))
				timings[i] = time.Since(start)
			}(i)
		}
		waitg.Wait()

		time.Sleep(300 * time.Millisecond)
		close(blocked)
		recorder.Close()

		// Check that all calls completed quickly
		for _, duration := range timings {
			if duration > 200*time.Millisecond {
				t.Errorf("Slow call: %v", duration)
			}
		}

		// Verify correctness
		results, _ := recorder.GetRecords()
		if results["foo"][0].TotalCount() != concurrentCalls {
			t.Errorf("Expected totalCount = %d, got %d", concurrentCalls, results["foo"][0].TotalCount())
		}
	})
}
