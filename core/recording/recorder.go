package recording

import (
	hdr "github.com/ofux/hdrhistogram"
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
	askForRecordsSnapshot chan chan<- RecordSnapshot
	recordingWaitGroup    *sync.WaitGroup
	processingWaitGroup   *sync.WaitGroup
}

type RecordSnapshot struct {
	HTTPRecordsOverTimeSnapshot *HTTPRecordsOverTimeSnapshot
	Err                         error
}

func NewRecorder(concurrent int) *Recorder {
	return &Recorder{
		recording:             READY,
		recordsQueue:          make(chan RecordEntry, concurrent),
		askForRecordsSnapshot: make(chan chan<- RecordSnapshot),
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

func (r *Recorder) processRecords(processRecord func(RecordEntry), processSnapshotRequest func(chan<- RecordSnapshot)) {
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
	// Max value represents one hour. Min value represents 1ms.
	return hdr.New(0, 3600*1000, 3)
}

func NanosecondToHistogramTime(nano int64) int64 {
	const nanoToHisto = 1000 * 1000 // Converts nanoseconds to milliseconds
	return nano / nanoToHisto
}

func mergeHistograms(h1, h2 *hdr.Histogram) *hdr.Histogram {
	h := h1.Copy()
	h.Merge(h2)
	return h
}
