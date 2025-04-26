package dto

type ChiiSubjectRec2 struct {
	SubjectID    int     `db:"subject_id"`
	RecSubjectID int     `db:"rec_subject_id"`
	MioSim       float64 `db:"mio_sim"`
	MioCount     int     `db:"mio_count"`
}
