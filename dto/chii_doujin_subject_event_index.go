package dto

type ChiiDoujinSubjectEventIndex struct {
	EventID     int `db:"event_id"`
	SubjectID   int `db:"subject_id"`
	SubjectType int `db:"subject_type"`
	Dateline    int `db:"dateline"`
}
