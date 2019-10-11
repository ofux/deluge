package recording

import (
	hdr "github.com/ofux/hdrhistogram"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestMapHTTPRecords(t *testing.T) {
	rec := buildHTTPRecordsOverTimeForTests(10, 20)
	mappedRec, err := MapHTTPRecords(rec)
	require.NoError(t, err)

	reMappedRec, err := MapPersistedHTTPRecords(mappedRec)
	require.NoError(t, err)

	assert.Equal(t, rec, reMappedRec)
}

func buildHTTPRecordsOverTimeForTests(concurrent, iterationCount int) *HTTPRecordsOverTime {
	records := &HTTPRecordsOverTime{
		Global:   buildHTTPRecordsForTests(concurrent),
		OverTime: make([]*HTTPRecord, 0, iterationCount),
	}

	for iter := 0; iter < iterationCount; iter++ {
		records.OverTime = append(records.OverTime, buildHTTPRecordsForTests(concurrent))
	}
	return records
}

func BenchmarkMapHTTPRecords(b *testing.B) {
	records := buildHTTPRecordsOverTimeForBenchmarks(1000, MaxOverTimeCount)

	for i := 0; i < b.N; i++ {
		_, err := MapHTTPRecords(records)

		if err != nil {
			b.Fatal(err)
		}
	}
}

func buildHTTPRecordsOverTimeForBenchmarks(concurrent, iterationCount int) *HTTPRecordsOverTime {
	records := &HTTPRecordsOverTime{
		Global: &HTTPRecord{
			HTTPRequestRecord: HTTPRequestRecord{
				Global:    createHistogram(),
				PerStatus: make(map[int]*hdr.Histogram),
				PerOkKo:   make(map[OkKo]*hdr.Histogram),
			},
			PerRequests: make(map[string]*HTTPRequestRecord),
		},
		OverTime: make([]*HTTPRecord, 0, iterationCount),
	}

	for iter := 0; iter < iterationCount; iter++ {
		records.OverTime = append(records.OverTime, createHTTPRecords(1)...)
		for i := 0; i < concurrent; i++ {
			rec := &HTTPRecordEntry{
				Iteration:  iter,
				Name:       "This is my awesome HTTP request",
				Value:      NanosecondToHistogramTime(1000*1000 + rand.Int63n(600*1000*1000*1000)),
				StatusCode: 200,
			}
			// Global record for all iterations
			processEntryToHTTPRecord(rec, records.Global)
			processEntryToHTTPRecord(rec, records.OverTime[iter])
		}
	}
	return records
}
