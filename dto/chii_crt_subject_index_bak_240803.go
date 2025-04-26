package dto

type ChiiCrtSubjectIndexBak240803 struct {
	CrtID         int    `db:"crt_id"`
	SubjectID     int    `db:"subject_id"`
	SubjectTypeID int    `db:"subject_type_id"`
	CrtType       int    `db:"crt_type"`
	CtrAppearEps  string `db:"ctr_appear_eps"`
	CrtOrder      int    `db:"crt_order"`
}
