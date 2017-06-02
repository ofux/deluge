package recording

import (
	"errors"
	hdr "github.com/codahale/hdrhistogram"
)

type HTTPRecorder struct {
	*Recorder
	histograms map[string][]*hdr.Histogram
}

type HTTPRecord struct {
	Iteration  int
	Name       string
	Value      int64
	StatusCode int
}

func NewHTTPRecorder(concurrent int) *HTTPRecorder {
	recorder := &HTTPRecorder{
		Recorder:   NewRecorder(concurrent),
		histograms: make(map[string][]*hdr.Histogram),
	}
	recorder.processRecords(recorder.processHTTPRecord)
	return recorder
}

func (r *HTTPRecorder) GetRecords() (map[string][]*hdr.Histogram, error) {
	if r.recording != TERMINATED {
		return nil, errors.New("Cannot get records while recording. Did you forget to call the 'Close()' method?")
	}
	return r.histograms, nil
}

func (r *HTTPRecorder) processHTTPRecord(record Record) {
	rec := record.(*HTTPRecord)
	histograms, ok := r.histograms[rec.Name]
	if !ok {
		histograms = make([]*hdr.Histogram, 0)
		r.histograms[rec.Name] = histograms
	}

	// TODO: optimize this
	if len(histograms) <= rec.Iteration {
		diff := rec.Iteration + 1 - len(histograms)
		histograms = append(histograms, createHistograms(diff)...)
		r.histograms[rec.Name] = histograms
	}

	histogram := histograms[rec.Iteration]
	histogram.RecordValue(rec.Value)
}

func createHistograms(count int) []*hdr.Histogram {
	histograms := make([]*hdr.Histogram, count)
	for i := 0; i < count; i++ {
		histograms[i] = hdr.New(0, 3600000000, 3)
	}
	return histograms
}
