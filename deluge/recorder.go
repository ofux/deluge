package deluge

import (
	"errors"
	hdr "github.com/codahale/hdrhistogram"
)

type RecordingState int

const (
	READY RecordingState = iota
	RECORDING
	TERMINATING
	TERMINATED
)

type Recorder interface {
	Record(iteration int, id string, value int64)
	Close()
	GetRecords() (map[string][]*hdr.Histogram, error)
}

type QueuedRecorder struct {
	recording    RecordingState
	recordsQueue chan Record
	histograms   map[string][]*hdr.Histogram
	onTerminated chan struct{}
}

type Record struct {
	iteration int
	id        string
	value     int64
}

func NewRecorder(concurrent int) Recorder {
	recorder := &QueuedRecorder{
		recording:    READY,
		recordsQueue: make(chan Record, concurrent),
		histograms:   make(map[string][]*hdr.Histogram),
		onTerminated: make(chan struct{}, 1),
	}
	recorder.processRecords()
	return recorder
}

// Record records a new value in the underlying appropriate HDRHistogram.
// This is safe to call this method from different goroutines.
// Calling this method on a closed Recorder will cause a panic.
func (r *QueuedRecorder) Record(iteration int, id string, value int64) {
	r.recordsQueue <- Record{
		iteration: iteration,
		id:        id,
		value:     value,
	}
}

// Close closes the Recorder, making the results available for read.
// Trying to record some values on a closed Recorder will cause a panic.
func (r *QueuedRecorder) Close() {
	if r.recording == RECORDING {
		r.recording = TERMINATING
		// ensure listener won't stay blocked
		close(r.recordsQueue)
		// wait for the end of recording
		<-r.onTerminated
		r.recording = TERMINATED
	}
}

func (r *QueuedRecorder) GetRecords() (map[string][]*hdr.Histogram, error) {
	if r.recording != TERMINATED {
		return nil, errors.New("Cannot get records while recording. Did you forget to call the 'Close()' method?")
	}
	return r.histograms, nil
}

func (r *QueuedRecorder) processRecords() {
	r.recording = RECORDING

	go func() {
		defer func() {
			r.onTerminated <- struct{}{}
		}()

		for {
			rec, ok := <-r.recordsQueue
			if !ok {
				return
			}

			histograms, ok := r.histograms[rec.id]
			if !ok {
				histograms = make([]*hdr.Histogram, 0)
				r.histograms[rec.id] = histograms
			}

			// TODO: optimize this
			if len(histograms) < rec.iteration+1 {
				histograms = append(histograms, createHistograms(rec.iteration+1-len(histograms))...)
				r.histograms[rec.id] = histograms
			}

			histogram := histograms[rec.iteration]
			histogram.RecordValue(rec.value)
		}
	}()

}

func createHistograms(count int) []*hdr.Histogram {
	histograms := make([]*hdr.Histogram, count)
	for i := 0; i < count; i++ {
		histograms[i] = hdr.New(0, 3600000000, 3)
	}
	return histograms
}
