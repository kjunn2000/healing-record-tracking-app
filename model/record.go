package model

import "time"

type Record struct {
	RecordId        int
	CaseId          int
	RecordName      string
	MetricUnit      string
	Details         []RecordDetail
	CreatedDateTime time.Time
}
