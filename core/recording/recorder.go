package recording

import (
	hdr "github.com/codahale/hdrhistogram"
	"github.com/ofux/deluge/repov2"
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
	recording             RecordingState
	recordsQueue          chan RecordEntry
	askForRecordsSnapshot chan chan<- *repov2.PersistedHTTPRecordsOverTime
	recordingWaitGroup    *sync.WaitGroup
	processingWaitGroup   *sync.WaitGroup
}

func NewRecorder(concurrent int) *Recorder {
	return &Recorder{
		recording:             READY,
		recordsQueue:          make(chan RecordEntry, concurrent),
		askForRecordsSnapshot: make(chan chan<- *repov2.PersistedHTTPRecordsOverTime),
		recordingWaitGroup:    new(sync.WaitGroup),
		processingWaitGroup:   new(sync.WaitGroup),
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

func (r *Recorder) processRecords(processRecord func(RecordEntry), processSnapshotRequest func(chan<- *repov2.PersistedHTTPRecordsOverTime)) {
	r.recording = RECORDING
	r.processingWaitGroup.Add(1)

	go func() {
		defer r.processingWaitGroup.Done()

		for {
			select {
			case rec, ok := <-r.recordsQueue:
				if !ok {
					return // exit for loop and goroutine when recordsQueue is closed
				}
				processRecord(rec)
			case snapshotChan, ok := <-r.askForRecordsSnapshot:
				if ok {
					processSnapshotRequest(snapshotChan)
				}
			}
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
