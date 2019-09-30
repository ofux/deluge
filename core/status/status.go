package status

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

// DelugeStatus represents the status of a running deluge (ie. a job)
type DelugeStatus int

const (
	// Order is important for merging. The highest wins.
	DelugeVirgin DelugeStatus = iota
	DelugeInProgress
	DelugeDoneSuccess
	DelugeInterrupted
	DelugeDoneError
)

func (s DelugeStatus) String() string {
	switch s {
	case DelugeVirgin:
		return "notStarted"
	case DelugeInProgress:
		return "inProgress"
	case DelugeDoneSuccess:
		return "doneSuccess"
	case DelugeInterrupted:
		return "interrupted"
	case DelugeDoneError:
		return "doneError"
	default:
		return "unknown"
	}
}

func (s DelugeStatus) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", t)
	return []byte(stamp), nil
}

func (s *DelugeStatus) UnmarshalJSON(data []byte) error {
	payload := strings.Trim(string(data), `"`)
	switch payload {
	case DelugeVirgin.String():
		*s = DelugeVirgin
	case DelugeInProgress.String():
		*s = DelugeInProgress
	case DelugeDoneSuccess.String():
		*s = DelugeDoneSuccess
	case DelugeInterrupted.String():
		*s = DelugeInterrupted
	case DelugeDoneError.String():
		*s = DelugeDoneError
	default:
		return errors.Errorf("invalid status '%s'", payload)
	}
	return nil
}

func MergeDelugeStatuses(s1, s2 DelugeStatus) DelugeStatus {
	if int(s1) > int(s2) {
		return s1
	}
	return s2
}

// ScenarioStatus represents the status of a running scenario
type ScenarioStatus int

const (
	// Order is important for merging. The highest wins.
	ScenarioVirgin ScenarioStatus = iota
	ScenarioInProgress
	ScenarioDoneSuccess
	ScenarioInterrupted
	ScenarioDoneError
)

func (s ScenarioStatus) String() string {
	switch s {
	case ScenarioVirgin:
		return "notStarted"
	case ScenarioInProgress:
		return "inProgress"
	case ScenarioDoneSuccess:
		return "doneSuccess"
	case ScenarioInterrupted:
		return "interrupted"
	case ScenarioDoneError:
		return "doneError"
	default:
		return "unknown"
	}
}

func (s ScenarioStatus) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", t)
	return []byte(stamp), nil
}

func (s *ScenarioStatus) UnmarshalJSON(data []byte) error {
	payload := strings.Trim(string(data), `"`)
	switch payload {
	case ScenarioVirgin.String():
		*s = ScenarioVirgin
	case ScenarioInProgress.String():
		*s = ScenarioInProgress
	case ScenarioDoneSuccess.String():
		*s = ScenarioDoneSuccess
	case ScenarioInterrupted.String():
		*s = ScenarioInterrupted
	case ScenarioDoneError.String():
		*s = ScenarioDoneError
	default:
		return errors.Errorf("invalid status '%s'", payload)
	}
	return nil
}

func MergeScenarioStatuses(s1, s2 ScenarioStatus) ScenarioStatus {
	if int(s1) > int(s2) {
		return s1
	}
	return s2
}
