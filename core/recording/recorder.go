package recording

import (
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

type OkKo string

const (
	Ok OkKo = "Ok"
	Ko OkKo = "Ko"
)

type RecordEntry interface{}

type Recorder struct {
	recording           RecordingState
	recordsQueue        chan RecordEntry
	recordingWaitGroup  *sync.WaitGroup
	processingWaitGroup *sync.WaitGroup
}

func NewRecorder() *Recorder {
	return &Recorder{
		recording:           READY,
		recordsQueue:        make(chan RecordEntry),
		recordingWaitGroup:  new(sync.WaitGroup),
		processingWaitGroup: new(sync.WaitGroup),
	}
}

// HTTPRecordEntry records a new Value in the underlying appropriate HDRHistogram.
// This is safe to call this method from different goroutines.
// Calling this method on a closed Recorder will cause a panic.
func (r *Recorder) Record(rec RecordEntry) {
	r.recordingWaitGroup.Add(1)
	go func() {
		defer r.recordingWaitGroup.Done()
		r.recordsQueue <- rec
	}()
}

// Close closes the Recorder, making the results available for read.
// Trying to record some values on a closed Recorder will cause a panic.
func (r *Recorder) Close() {
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

func (r *Recorder) processRecords(processRecord func(RecordEntry)) {
	r.recording = RECORDING
	r.processingWaitGroup.Add(1)

	go func() {
		defer r.processingWaitGroup.Done()

		for {
			rec, ok := <-r.recordsQueue
			if !ok {
				return
			}
			processRecord(rec)
		}
	}()

}

func createHistogram() *hdr.Histogram {
	return hdr.New(0, 3600000000, 3)
}

func mergeHistograms(h1, h2 *hdr.Histogram) *hdr.Histogram {
	h := hdr.Import(h1.Export())
	h.Merge(h2)
	return h
}

func copyHistogram(h *hdr.Histogram) *hdr.Histogram {
	return hdr.Import(h.Export())
}
