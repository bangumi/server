package dto

type ChiiPersonCsIndex struct {
	PrsnType      string `db:"prsn_type"`
	PrsnID        int    `db:"prsn_id"`
	PrsnPosition  int    `db:"prsn_position"`
	SubjectID     int    `db:"subject_id"`
	SubjectTypeID int    `db:"subject_type_id"`
	Summary       string `db:"summary"`
	PrsnAppearEps string `db:"prsn_appear_eps"`
}
