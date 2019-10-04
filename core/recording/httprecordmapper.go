package recording

import (
	"github.com/ofux/deluge/repov2"
	hdr "github.com/ofux/hdrhistogram"
)

func MapHTTPRecords(records *HTTPRecordsOverTime) (*repov2.PersistedHTTPRecordsOverTime, error) {
	p, err := mapHTTPRecord(records.Global)
	if err != nil {
		return nil, err
	}
	report := &repov2.PersistedHTTPRecordsOverTime{
		Global:   p,
		OverTime: make([]*repov2.PersistedHTTPRecord, 0, len(records.OverTime)),
	}
	for _, v := range records.OverTime {
		p, err := mapHTTPRecord(v)
		if err != nil {
			return nil, err
		}
		report.OverTime = append(report.OverTime, p)
	}

	return report, nil
}

func mapHTTPRecord(rec *HTTPRecord) (*repov2.PersistedHTTPRecord, error) {
	p, err := mapHTTPRequestRecord(&(rec.HTTPRequestRecord))
	if err != nil {
		return nil, err
	}
	st := &repov2.PersistedHTTPRecord{
		PersistedHTTPRequestRecord: *p,
		PerRequests:                make(map[string]*repov2.PersistedHTTPRequestRecord),
	}
	for k, v := range rec.PerRequests {
		p, err := mapHTTPRequestRecord(v)
		if err != nil {
			return nil, err
		}
		st.PerRequests[k] = p
	}
	return st, nil
}

func mapHTTPRequestRecord(rec *HTTPRequestRecord) (*repov2.PersistedHTTPRequestRecord, error) {
	snap, err := rec.Global.Export()
	if err != nil {
		return nil, err
	}
	st := &repov2.PersistedHTTPRequestRecord{
		Global:    snap,
		PerStatus: make(map[int]*hdr.Snapshot),
		PerOkKo:   make(map[repov2.OkKo]*hdr.Snapshot),
	}
	for k, v := range rec.PerStatus {
		snap, err := v.Export()
		if err != nil {
			return nil, err
		}
		st.PerStatus[k] = snap
	}
	for k, v := range rec.PerOkKo {
		key := repov2.Ok
		if k == Ko {
			key = repov2.Ko
		}
		snap, err := v.Export()
		if err != nil {
			return nil, err
		}
		st.PerOkKo[key] = snap
	}
	return st, nil
}

func MapPersistedHTTPRecords(records *repov2.PersistedHTTPRecordsOverTime) (*HTTPRecordsOverTime, error) {
	if records == nil {
		return nil, nil
	}
	p, err := mapPersistedHTTPRecord(records.Global)
	if err != nil {
		return nil, err
	}
	report := &HTTPRecordsOverTime{
		Global:   p,
		OverTime: make([]*HTTPRecord, 0, len(records.OverTime)),
	}
	for _, v := range records.OverTime {
		p, err := mapPersistedHTTPRecord(v)
		if err != nil {
			return nil, err
		}
		report.OverTime = append(report.OverTime, p)
	}

	return report, nil
}

func mapPersistedHTTPRecord(rec *repov2.PersistedHTTPRecord) (*HTTPRecord, error) {
	p, err := mapPersistedHTTPRequestRecord(&(rec.PersistedHTTPRequestRecord))
	if err != nil {
		return nil, err
	}
	st := &HTTPRecord{
		HTTPRequestRecord: *p,
		PerRequests:       make(map[string]*HTTPRequestRecord),
	}
	for k, v := range rec.PerRequests {
		p, err := mapPersistedHTTPRequestRecord(v)
		if err != nil {
			return nil, err
		}
		st.PerRequests[k] = p
	}
	return st, nil
}

func mapPersistedHTTPRequestRecord(rec *repov2.PersistedHTTPRequestRecord) (*HTTPRequestRecord, error) {
	h, err := hdr.Import(rec.Global)
	if err != nil {
		return nil, err
	}
	st := &HTTPRequestRecord{
		Global:    h,
		PerStatus: make(map[int]*hdr.Histogram),
		PerOkKo:   make(map[OkKo]*hdr.Histogram),
	}
	for k, v := range rec.PerStatus {
		h, err := hdr.Import(v)
		if err != nil {
			return nil, err
		}
		st.PerStatus[k] = h
	}
	for k, v := range rec.PerOkKo {
		key := Ok
		if k == repov2.Ko {
			key = Ko
		}
		h, err := hdr.Import(v)
		if err != nil {
			return nil, err
		}
		st.PerOkKo[key] = h
	}
	return st, nil
}
