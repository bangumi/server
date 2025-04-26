package dto

type ChiiCrtCastIndex struct {
	CrtID         int    `db:"crt_id"`
	PrsnID        int    `db:"prsn_id"`
	SubjectID     int    `db:"subject_id"`
	SubjectTypeID int    `db:"subject_type_id"`
	Summary       string `db:"summary"`
}
