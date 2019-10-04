package recording

import (
	hdr "github.com/ofux/hdrhistogram"
)

func copyHTTPRecord(rec *HTTPRecord) *HTTPRecord {
	st := &HTTPRecord{
		HTTPRequestRecord: *copyHTTPRequestRecord(&(rec.HTTPRequestRecord)),
		PerRequests:       make(map[string]*HTTPRequestRecord),
	}
	for k, v := range rec.PerRequests {
		st.PerRequests[k] = copyHTTPRequestRecord(v)
	}
	return st
}

func copyHTTPRequestRecord(rec *HTTPRequestRecord) *HTTPRequestRecord {
	st := &HTTPRequestRecord{
		Global:    rec.Global.Copy(),
		PerStatus: make(map[int]*hdr.Histogram),
		PerOkKo:   make(map[OkKo]*hdr.Histogram),
	}
	for k, v := range rec.PerStatus {
		st.PerStatus[k] = v.Copy()
	}
	for k, v := range rec.PerOkKo {
		st.PerOkKo[k] = v.Copy()
	}
	return st
}
