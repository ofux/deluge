package deluge

import (
	"errors"
	hdr "github.com/codahale/hdrhistogram"
	"sync"
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
	recording           RecordingState
	recordsQueue        chan Record
	recordingWaitGroup  *sync.WaitGroup
	processingWaitGroup *sync.WaitGroup
	histograms          map[string][]*hdr.Histogram
}

type Record struct {
	iteration int
	id        string
	value     int64
}

func NewRecorder(concurrent int) Recorder {
	recorder := &QueuedRecorder{
		recording:           READY,
		recordsQueue:        make(chan Record, concurrent),
		recordingWaitGroup:  new(sync.WaitGroup),
		processingWaitGroup: new(sync.WaitGroup),
		histograms:          make(map[string][]*hdr.Histogram),
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

func (r *QueuedRecorder) GetRecords() (map[string][]*hdr.Histogram, error) {
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
				histograms = make([]*hdr.Histogram, 0)
				r.histograms[rec.id] = histograms
			}

			// TODO: optimize this
			if len(histograms) <= rec.iteration {
				diff := rec.iteration + 1 - len(histograms)
				histograms = append(histograms, createHistograms(diff)...)
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
