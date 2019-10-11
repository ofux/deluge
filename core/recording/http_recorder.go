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
	records                              *HTTPRecordsOverTime
	affectedTimeIndexesSinceLastSnapshot map[int]struct{}
	overTimeCount                        int
	iterationCount                       int
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

type HTTPRecordsOverTimeSnapshot struct {
	Global   *HTTPRecord
	OverTime map[int]*HTTPRecord
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
		affectedTimeIndexesSinceLastSnapshot: make(map[int]struct{}),
		iterationCount:                       iterationCount,
		overTimeCount:                        overTimeCount,
	}
	recorder.processRecords(recorder.processHTTPEntry, recorder.processRecordsSnapshotRequest)
	return recorder
}

// GetRecords returns the full records and can be called only once recording has ended.
func (r *HTTPRecorder) GetRecords() (*HTTPRecordsOverTime, error) {
	if r.recording != TERMINATED {
		return nil, errors.New("GetRecords can only be called after recording ended properly and after the 'Close()' method has been called")
	}
	return r.records, nil
}

// GetRecordsSnapshot returns a channel where a copy of current records will be sent.
func (r *HTTPRecorder) GetRecordsSnapshot() (<-chan RecordSnapshot, error) {
	if r.recording != RECORDING {
		return nil, errors.New("GetRecordsSnapshot can only be called while recording. Use GetRecords instead")
	}
	// We set a buffer of size 1 so 'processRecordsSnapshotRequest' can never stay blocked (waiting for a listener)
	newChan := make(chan RecordSnapshot, 1)
	r.askForRecordsSnapshot <- newChan
	return newChan, nil
}

func (r *HTTPRecorder) processRecordsSnapshotRequest(snapshotChan chan<- RecordSnapshot) {
	snap := &HTTPRecordsOverTimeSnapshot{
		Global:   copyHTTPRecord(r.records.Global),
		OverTime: make(map[int]*HTTPRecord),
	}
	for index := range r.affectedTimeIndexesSinceLastSnapshot {
		snap.OverTime[index] = copyHTTPRecord(r.records.OverTime[index])
	}

	// Clear affectedTimeIndexesSinceLastSnapshot map
	r.affectedTimeIndexesSinceLastSnapshot = make(map[int]struct{})

	snapshotChan <- RecordSnapshot{
		HTTPRecordsOverTimeSnapshot: snap,
		Err:                         nil,
	}
}

func (r *HTTPRecorder) processHTTPEntry(record RecordEntry) {
	rec := record.(*HTTPRecordEntry)

	// Global record for all iterations
	processEntryToHTTPRecord(rec, r.records.Global)

	overTimeIndex := r.iterationToTimeIndex(rec.Iteration)
	if len(r.records.OverTime) <= overTimeIndex {
		diff := overTimeIndex + 1 - len(r.records.OverTime)
		r.records.OverTime = append(r.records.OverTime, createHTTPRecords(diff)...)
	}
	processEntryToHTTPRecord(rec, r.records.OverTime[overTimeIndex])
	r.affectedTimeIndexesSinceLastSnapshot[overTimeIndex] = struct{}{}
}

func (r *HTTPRecorder) iterationToTimeIndex(iteration int) int {
	return iteration * r.overTimeCount / r.iterationCount
}

func processEntryToHTTPRecord(rec *HTTPRecordEntry, out *HTTPRecord) {

	val := rec.Value
	if val < out.Global.LowestTrackableValue() {
		val = out.Global.LowestTrackableValue()
	}
	if val > out.Global.HighestTrackableValue() {
		val = out.Global.HighestTrackableValue()
	}

	// Global. We explicitly ignore the error as we already made sure 'val' is trackable
	_ = out.Global.RecordValue(val)

	// Global per status
	histogram, ok := out.PerStatus[rec.StatusCode]
	if !ok {
		histogram = createHistogram()
		out.PerStatus[rec.StatusCode] = histogram
	}
	_ = histogram.RecordValue(val)

	// Global per result OK/KO
	histogram, ok = out.PerOkKo[httpOkKo(rec)]
	if !ok {
		histogram = createHistogram()
		out.PerOkKo[httpOkKo(rec)] = histogram
	}
	_ = histogram.RecordValue(val)

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
	_ = requestRecords.Global.RecordValue(val)

	// Global per status
	histogram, ok = requestRecords.PerStatus[rec.StatusCode]
	if !ok {
		histogram = createHistogram()
		requestRecords.PerStatus[rec.StatusCode] = histogram
	}
	_ = histogram.RecordValue(val)

	// Global per result OK/KO
	histogram, ok = requestRecords.PerOkKo[httpOkKo(rec)]
	if !ok {
		histogram = createHistogram()
		requestRecords.PerOkKo[httpOkKo(rec)] = histogram
	}
	_ = histogram.RecordValue(val)
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
