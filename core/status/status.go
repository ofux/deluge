package status

type DelugeStatus int

const (
	// Order is important for merging statuses. The highest wins.
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

type ScenarioStatus int

const (
	ScenarioVirgin ScenarioStatus = iota
	ScenarioInProgress
	ScenarioDoneSuccess
	ScenarioDoneError
	ScenarioInterrupted
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

func MergeDelugeStatuses(s1, s2 DelugeStatus) DelugeStatus {
	if int(s1) > int(s2) {
		return s1
	}
	return s2
}

func MergeScenarioStatuses(s1, s2 ScenarioStatus) ScenarioStatus {
	if int(s1) > int(s2) {
		return s1
	}
	return s2
}
