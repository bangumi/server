package dto

type ChiiDoujinSubjectCollects struct {
	CltUID         int `db:"clt_uid"`
	CltSubjectID   int `db:"clt_subject_id"`
	CltSubjectType int `db:"clt_subject_type"`
	CltDateline    int `db:"clt_dateline"`
}
