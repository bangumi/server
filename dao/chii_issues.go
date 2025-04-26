package dto

type ChiiIssues struct {
	IsuID       int    `db:"isu_id"`
	IsuType     int    `db:"isu_type"`
	IsuMainID   int    `db:"isu_main_id"`
	IsuValue    int    `db:"isu_value"`
	IsuCreator  int    `db:"isu_creator"`
	IsuOperator int    `db:"isu_operator"`
	IsuStatus   int    `db:"isu_status"`
	IsuReason   string `db:"isu_reason"`
	IsuRelated  int    `db:"isu_related"`
	IsuDateline int    `db:"isu_dateline"`
}
