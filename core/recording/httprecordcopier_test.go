package recording

import (
	hdr "github.com/ofux/hdrhistogram"
	"reflect"
	"testing"
)

func Test_copyHTTPRecord(t *testing.T) {
	record := buildHTTPRecordsForTests(100)
	recordCopy := copyHTTPRecord(record)
	if !reflect.DeepEqual(recordCopy, record) {
		t.Errorf("copyHTTPRecord() = %v, want %v", recordCopy, record)
	}
}

func Benchmark_copyHTTPRecord(b *testing.B) {
	record := buildHTTPRecordsForTests(100)

	for i := 0; i < b.N; i++ {
		copyHTTPRecord(record)
	}
}

func buildHTTPRecordsForTests(concurrent int) *HTTPRecord {
	records := &HTTPRecord{
		HTTPRequestRecord: HTTPRequestRecord{
			Global:    createHistogram(),
			PerStatus: make(map[int]*hdr.Histogram),
			PerOkKo:   make(map[OkKo]*hdr.Histogram),
		},
		PerRequests: make(map[string]*HTTPRequestRecord),
	}

	for i := 0; i < concurrent; i++ {
		rec := &HTTPRecordEntry{
			Iteration:  42,
			Name:       "This is my awesome HTTP request",
			Value:      NanosecondToHistogramTime(int64(1000 * 1000 * i)),
			StatusCode: 200,
		}
		processEntryToHTTPRecord(rec, records)
		rec = &HTTPRecordEntry{
			Iteration:  42,
			Name:       "This is my other awesome HTTP request",
			Value:      NanosecondToHistogramTime(int64(1000 * 1000 * i * 3)),
			StatusCode: 401,
		}
		processEntryToHTTPRecord(rec, records)
		rec = &HTTPRecordEntry{
			Iteration:  41,
			Name:       "This is my other other awesome HTTP request",
			Value:      NanosecondToHistogramTime(int64(1000 * 1000 * i * 2)),
			StatusCode: 500,
		}
		processEntryToHTTPRecord(rec, records)
	}
	return records
}
