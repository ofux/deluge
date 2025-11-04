Let's define the following interface:
type Recorder interface {
    Record(iteration int, id string, value int64)
    Close()
    GetRecords() (map[string][]*hdr.Histogram, error)
}

Add the interface and implement it using HDRHistogram. Update the simulation to perform HTTP calls and record them with id equal to:
"request_name->status_code". The HDR Histograms must have a max value of 1 hour and a min value of 0, with 3 significant digits.
The implementation must be instantiable by calling `NewRecorder() Recorder`.
`NewSimUser` must now accept a `Recorder` as a parameter: `NewSimUser(name string, script ast.Node, recorder Recorder) *SimUser`.

Use the following dependency for HDRHistograms: `github.com/ofux/hdrhistogram v0.0.0-20191004163134-d8ebd25fd911`
