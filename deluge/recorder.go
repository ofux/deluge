package deluge

import (
	"sync"

	hdr "github.com/ofux/hdrhistogram"
)

type RecordingState int

const (
	READY RecordingState = iota
	RECORDING
	TERMINATING
	TERMINATED
)

// Histogram is an interface for recording and retrieving histogram data.
// It abstracts the underlying histogram implementation.
type Histogram interface {
	RecordValue(value int64) error
	TotalCount() int64
}

// HistogramConstructor is a function type that creates new Histogram instances.
type HistogramConstructor func() Histogram

type Recorder interface {
	Record(iteration int, id string, value int64)
	Close()
	GetRecords() (map[string][]Histogram, error)
}

// DefaultHistogramConstructor returns a HistogramConstructor that creates
// HDR histograms with default settings (min=0, max=3600000000, sigfigs=3).
func DefaultHistogramConstructor() HistogramConstructor {
	return func() Histogram {
		return hdr.New(0, 3600000000, 3)
	}
}

type HDRRecorder struct {
	mu                   sync.Mutex
	records              map[string][]Histogram
	histogramConstructor HistogramConstructor
}

func NewRecorder(histogramConstructor HistogramConstructor) Recorder {
	return &HDRRecorder{
		records:              make(map[string][]Histogram),
		histogramConstructor: histogramConstructor,
	}
}

func (r *HDRRecorder) Record(iteration int, id string, value int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.records[id]; !ok {
		r.records[id] = make([]Histogram, 0)
	}

	for len(r.records[id]) <= iteration {
		r.records[id] = append(r.records[id], r.histogramConstructor())
	}

	r.records[id][iteration].RecordValue(value)
}

func (r *HDRRecorder) Close() {
}

func (r *HDRRecorder) GetRecords() (map[string][]Histogram, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.records, nil
}
