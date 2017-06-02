package recording

import (
	"sync"
)

type RecordingState int

const (
	READY RecordingState = iota
	RECORDING
	TERMINATING
	TERMINATED
)

type Record interface{}

type Recorder struct {
	recording           RecordingState
	recordsQueue        chan Record
	recordingWaitGroup  *sync.WaitGroup
	processingWaitGroup *sync.WaitGroup
}

func NewRecorder(concurrent int) *Recorder {
	return &Recorder{
		recording:           READY,
		recordsQueue:        make(chan Record, concurrent),
		recordingWaitGroup:  new(sync.WaitGroup),
		processingWaitGroup: new(sync.WaitGroup),
	}
}

// HTTPRecord records a new Value in the underlying appropriate HDRHistogram.
// This is safe to call this method from different goroutines.
// Calling this method on a closed Recorder will cause a panic.
func (r *Recorder) Record(rec Record) {
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

func (r *Recorder) processRecords(processRecord func(Record)) {
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
