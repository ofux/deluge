package recording

import (
	hdr "github.com/codahale/hdrhistogram"
	"github.com/ofux/deluge/repov2"
)

func MapHTTPRecords(records *HTTPRecordsOverTime) *repov2.PersistedHTTPRecordsOverTime {
	report := &repov2.PersistedHTTPRecordsOverTime{
		Global:   mapHTTPRecord(records.Global),
		OverTime: make([]*repov2.PersistedHTTPRecord, 0, 16),
	}
	for _, v := range records.OverTime {
		report.OverTime = append(report.OverTime, mapHTTPRecord(v))
	}

	return report
}

func mapHTTPRecord(rec *HTTPRecord) *repov2.PersistedHTTPRecord {
	st := &repov2.PersistedHTTPRecord{
		PersistedHTTPRequestRecord: *mapHTTPRequestRecord(&(rec.HTTPRequestRecord)),
		PerRequests:                make(map[string]*repov2.PersistedHTTPRequestRecord),
	}
	for k, v := range rec.PerRequests {
		st.PerRequests[k] = mapHTTPRequestRecord(v)
	}
	return st
}

func mapHTTPRequestRecord(rec *HTTPRequestRecord) *repov2.PersistedHTTPRequestRecord {
	st := &repov2.PersistedHTTPRequestRecord{
		Global:    rec.Global.Export(),
		PerStatus: make(map[int]*hdr.Snapshot),
		PerOkKo:   make(map[repov2.OkKo]*hdr.Snapshot),
	}
	for k, v := range rec.PerStatus {
		st.PerStatus[k] = v.Export()
	}
	for k, v := range rec.PerOkKo {
		key := repov2.Ok
		if k == Ko {
			key = repov2.Ko
		}
		st.PerOkKo[key] = v.Export()
	}
	return st
}

func MapPersistedHTTPRecords(records *repov2.PersistedHTTPRecordsOverTime) *HTTPRecordsOverTime {
	if records == nil {
		return nil
	}
	report := &HTTPRecordsOverTime{
		Global:   mapPersistedHTTPRecord(records.Global),
		OverTime: make([]*HTTPRecord, 0, len(records.OverTime)),
	}
	for _, v := range records.OverTime {
		report.OverTime = append(report.OverTime, mapPersistedHTTPRecord(v))
	}

	return report
}

func mapPersistedHTTPRecord(rec *repov2.PersistedHTTPRecord) *HTTPRecord {
	st := &HTTPRecord{
		HTTPRequestRecord: *mapPersistedHTTPRequestRecord(&(rec.PersistedHTTPRequestRecord)),
		PerRequests:       make(map[string]*HTTPRequestRecord),
	}
	for k, v := range rec.PerRequests {
		st.PerRequests[k] = mapPersistedHTTPRequestRecord(v)
	}
	return st
}

func mapPersistedHTTPRequestRecord(rec *repov2.PersistedHTTPRequestRecord) *HTTPRequestRecord {
	st := &HTTPRequestRecord{
		Global:    hdr.Import(rec.Global),
		PerStatus: make(map[int]*hdr.Histogram),
		PerOkKo:   make(map[OkKo]*hdr.Histogram),
	}
	for k, v := range rec.PerStatus {
		st.PerStatus[k] = hdr.Import(v)
	}
	for k, v := range rec.PerOkKo {
		key := Ok
		if k == repov2.Ko {
			key = Ko
		}
		st.PerOkKo[key] = hdr.Import(v)
	}
	return st
}
