package dto

type ChiiSubjectRec struct {
	SubjectID    int     `db:"subject_id"`
	RecSubjectID int     `db:"rec_subject_id"`
	MioSim       float64 `db:"mio_sim"`
	MioCount     int     `db:"mio_count"`
}
