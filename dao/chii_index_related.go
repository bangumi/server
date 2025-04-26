package dto

type ChiiIndexRelated struct {
	IDxRltID       int    `db:"idx_rlt_id"`
	IDxRltCat      int    `db:"idx_rlt_cat"`
	IDxRltRid      int    `db:"idx_rlt_rid"`
	IDxRltType     int    `db:"idx_rlt_type"`
	IDxRltSid      int    `db:"idx_rlt_sid"`
	IDxRltOrder    int    `db:"idx_rlt_order"`
	IDxRltAward    string `db:"idx_rlt_award"`
	IDxRltComment  string `db:"idx_rlt_comment"`
	IDxRltDateline int    `db:"idx_rlt_dateline"`
	IDxRltBan      int    `db:"idx_rlt_ban"`
}
