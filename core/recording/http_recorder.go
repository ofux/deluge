package recording

import (
	"errors"
	hdr "github.com/codahale/hdrhistogram"
	"github.com/ofux/deluge/repov2"
)

type HTTPRecorder struct {
	*Recorder
	records *HTTPRecordsOverTime
}

type HTTPRecordsOverTime struct {
	Global       *HTTPRecord
	PerIteration []*HTTPRecord
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

func NewHTTPRecorder(concurrent int) *HTTPRecorder {
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
			PerIteration: make([]*HTTPRecord, 0, 16),
		},
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
func (r *HTTPRecorder) GetRecordsSnapshot() (<-chan *repov2.PersistedHTTPRecordsOverTime, error) {
	if r.recording != RECORDING {
		return nil, errors.New("GetRecordsSnapshot should be used while recording. Used GetRecords instead")
	}
	newChan := make(chan *repov2.PersistedHTTPRecordsOverTime, 1)
	r.askForRecordsSnapshot <- newChan
	return newChan, nil
}

func (r *HTTPRecorder) processRecordsSnapshotRequest(snapshotChan chan<- *repov2.PersistedHTTPRecordsOverTime) {
	snapshotChan <- MapHTTPRecords(r.records)
}

func (r *HTTPRecorder) processHTTPEntry(record RecordEntry) {
	rec := record.(*HTTPRecordEntry)

	// Global record for all iterations
	r.processEntryToHTTPRecord(rec, r.records.Global)

	if len(r.records.PerIteration) <= rec.Iteration {
		diff := rec.Iteration + 1 - len(r.records.PerIteration)
		r.records.PerIteration = append(r.records.PerIteration, createHTTPRecords(diff)...)
	}
	r.processEntryToHTTPRecord(rec, r.records.PerIteration[rec.Iteration])
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
