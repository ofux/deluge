package deluge

import (
	"errors"
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

type QueuedRecorder struct {
	recording           RecordingState
	recordsQueue        chan Record
	recordingWaitGroup  *sync.WaitGroup
	processingWaitGroup *sync.WaitGroup
	histograms          map[string][]Histogram
	newHistogram        HistogramConstructor
}

type Record struct {
	iteration int
	id        string
	value     int64
}

func NewRecorder(newHistogram HistogramConstructor) Recorder {
	recorder := &QueuedRecorder{
		recording:           READY,
		recordsQueue:        make(chan Record, 10),
		recordingWaitGroup:  new(sync.WaitGroup),
		processingWaitGroup: new(sync.WaitGroup),
		histograms:          make(map[string][]Histogram),
		newHistogram:        newHistogram,
	}
	recorder.processRecords()
	return recorder
}

// Record records a new value in the underlying appropriate HDRHistogram.
// This is safe to call this method from different goroutines.
// Calling this method on a closed Recorder will cause a panic.
func (r *QueuedRecorder) Record(iteration int, id string, value int64) {
	r.recordingWaitGroup.Add(1)
	go func() {
		defer r.recordingWaitGroup.Done()
		r.recordsQueue <- Record{
			iteration: iteration,
			id:        id,
			value:     value,
		}
	}()
}

// Close closes the Recorder, making the results available for read.
// Trying to record some values on a closed Recorder will cause a panic.
func (r *QueuedRecorder) Close() {
	if r.recording == RECORDING {
		r.recording = TERMINATING
		// wait for all records to be taken
		r.recordingWaitGroup.Wait()
		// ensure listener won't stay blocked
		close(r.recordsQueue)
		// wait for the end of recording
		r.processingWaitGroup.Wait()
		r.recording = TERMINATED
	}
}

func (r *QueuedRecorder) GetRecords() (map[string][]Histogram, error) {
	if r.recording != TERMINATED {
		return nil, errors.New("Cannot get records while recording. Did you forget to call the 'Close()' method?")
	}
	return r.histograms, nil
}

func (r *QueuedRecorder) processRecords() {
	r.recording = RECORDING
	r.processingWaitGroup.Add(1)

	go func() {
		defer r.processingWaitGroup.Done()

		for {
			rec, ok := <-r.recordsQueue
			if !ok {
				return
			}

			histograms, ok := r.histograms[rec.id]
			if !ok {
				histograms = make([]Histogram, 0)
				r.histograms[rec.id] = histograms
			}

			// TODO: optimize this
			if len(histograms) <= rec.iteration {
				diff := rec.iteration + 1 - len(histograms)
				histograms = append(histograms, r.createHistograms(diff)...)
				r.histograms[rec.id] = histograms
			}

			histogram := histograms[rec.iteration]
			histogram.RecordValue(rec.value)
		}
	}()

}

func (r *QueuedRecorder) createHistograms(count int) []Histogram {
	histograms := make([]Histogram, count)
	for i := 0; i < count; i++ {
		histograms[i] = r.newHistogram()
	}
	return histograms
}

// DefaultHistogramConstructor returns a HistogramConstructor that creates
// HDR histograms with default settings (min=0, max=3600000000, sigfigs=3).
func DefaultHistogramConstructor() HistogramConstructor {
	return func() Histogram {
		return hdr.New(0, 3600000000, 3)
	}
}
