package recording

import (
	"errors"
	hdr "github.com/ofux/hdrhistogram"
)

const (
	MaxOverTimeCount int = 360
)

type HTTPRecorder struct {
	*Recorder
	records        *HTTPRecordsOverTime
	overTimeCount  int
	iterationCount int
}

type HTTPRecordsOverTime struct {
	Global   *HTTPRecord
	OverTime []*HTTPRecord
}

type HTTPRecord struct {
	HTTPRequestRecord
	PerRequests map[string]*HTTPRequestRecord
}

type HTTPRequestRecord struct {
	Global    *hdr.Histogram
	PerStatus map[int]*hdr.Histogram
	PerOkKo   map[OkKo]*hdr.Histogram
}

type HTTPRecordEntry struct {
	Iteration  int
	Name       string
	Value      int64
	StatusCode int
}

func NewHTTPRecorder(iterationCount, concurrent int) *HTTPRecorder {
	overTimeCount := Min(iterationCount, MaxOverTimeCount)

	recorder := &HTTPRecorder{
		Recorder: NewRecorder(concurrent),
		records: &HTTPRecordsOverTime{
			Global: &HTTPRecord{
				HTTPRequestRecord: HTTPRequestRecord{
					Global:    createHistogram(),
					PerStatus: make(map[int]*hdr.Histogram),
					PerOkKo:   make(map[OkKo]*hdr.Histogram),
				},
				PerRequests: make(map[string]*HTTPRequestRecord),
			},
			OverTime: make([]*HTTPRecord, 0, overTimeCount),
		},
		iterationCount: iterationCount,
		overTimeCount:  overTimeCount,
	}
	recorder.processRecords(recorder.processHTTPEntry, recorder.processRecordsSnapshotRequest)
	return recorder
}

// GetRecords returns the full records and can be called only once recording has ended.
func (r *HTTPRecorder) GetRecords() (*HTTPRecordsOverTime, error) {
	if r.recording != TERMINATED {
		return nil, errors.New("Cannot get records while recording. Did you forget to call the 'Close()' method?")
	}
	return r.records, nil
}

// GetRecordsSnapshot returns a channel where a copy of current records will be sent.
func (r *HTTPRecorder) GetRecordsSnapshot() (<-chan RecordSnapshot, error) {
	if r.recording != RECORDING {
		return nil, errors.New("GetRecordsSnapshot should be used while recording. Used GetRecords instead")
	}
	newChan := make(chan RecordSnapshot, 1)
	r.askForRecordsSnapshot <- newChan
	return newChan, nil
}

func (r *HTTPRecorder) processRecordsSnapshotRequest(snapshotChan chan<- RecordSnapshot) {
	snap, err := MapHTTPRecords(r.records)
	snapshotChan <- RecordSnapshot{
		HTTPRecordsOverTime: snap,
		Err:                 err,
	}
}

func (r *HTTPRecorder) processHTTPEntry(record RecordEntry) {
	rec := record.(*HTTPRecordEntry)

	// Global record for all iterations
	r.processEntryToHTTPRecord(rec, r.records.Global)

	overTimeIndex := r.iterationToTimeIndex(rec.Iteration)
	if len(r.records.OverTime) <= overTimeIndex {
		diff := overTimeIndex + 1 - len(r.records.OverTime)
		r.records.OverTime = append(r.records.OverTime, createHTTPRecords(diff)...)
	}
	r.processEntryToHTTPRecord(rec, r.records.OverTime[overTimeIndex])
}

func (r *HTTPRecorder) iterationToTimeIndex(iteration int) int {
	return iteration * r.overTimeCount / r.iterationCount
}

func (r *HTTPRecorder) processEntryToHTTPRecord(rec *HTTPRecordEntry, out *HTTPRecord) {

	// Global
	out.Global.RecordValue(rec.Value)

	// Global per status
	histogram, ok := out.PerStatus[rec.StatusCode]
	if !ok {
		histogram = createHistogram()
		out.PerStatus[rec.StatusCode] = histogram
	}
	histogram.RecordValue(rec.Value)

	// Global per result OK/KO
	histogram, ok = out.PerOkKo[httpOkKo(rec)]
	if !ok {
		histogram = createHistogram()
		out.PerOkKo[httpOkKo(rec)] = histogram
	}
	histogram.RecordValue(rec.Value)

	// Request's records
	requestRecords, ok := out.PerRequests[rec.Name]
	if !ok {
		requestRecords = &HTTPRequestRecord{
			Global:    createHistogram(),
			PerStatus: make(map[int]*hdr.Histogram),
			PerOkKo:   make(map[OkKo]*hdr.Histogram),
		}
		out.PerRequests[rec.Name] = requestRecords
	}

	// Request's global
	requestRecords.Global.RecordValue(rec.Value)

	// Global per status
	histogram, ok = requestRecords.PerStatus[rec.StatusCode]
	if !ok {
		histogram = createHistogram()
		requestRecords.PerStatus[rec.StatusCode] = histogram
	}
	histogram.RecordValue(rec.Value)

	// Global per result OK/KO
	histogram, ok = requestRecords.PerOkKo[httpOkKo(rec)]
	if !ok {
		histogram = createHistogram()
		requestRecords.PerOkKo[httpOkKo(rec)] = histogram
	}
	histogram.RecordValue(rec.Value)
}

func createHTTPRecords(count int) []*HTTPRecord {
	httpRecords := make([]*HTTPRecord, count)
	for i := 0; i < count; i++ {
		httpRecords[i] = &HTTPRecord{
			HTTPRequestRecord: HTTPRequestRecord{
				Global:    createHistogram(),
				PerStatus: make(map[int]*hdr.Histogram),
				PerOkKo:   make(map[OkKo]*hdr.Histogram),
			},
			PerRequests: make(map[string]*HTTPRequestRecord),
		}
	}
	return httpRecords
}

func httpOkKo(httpRec *HTTPRecordEntry) OkKo {
	if httpRec.StatusCode < 400 {
		return Ok
	}
	return Ko
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
