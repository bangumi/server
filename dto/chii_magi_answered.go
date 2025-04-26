package dto

type ChiiMagiAnswered struct {
	AswQID      int `db:"asw_qid"`
	AswUID      int `db:"asw_uid"`
	AswAnswer   int `db:"asw_answer"`
	AswResult   int `db:"asw_result"`
	AswDateline int `db:"asw_dateline"`
}
