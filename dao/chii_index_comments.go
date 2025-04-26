package dto

type ChiiIndexComments struct {
	IDxPstID       int    `db:"idx_pst_id"`
	IDxPstMid      int    `db:"idx_pst_mid"`
	IDxPstUid      int    `db:"idx_pst_uid"`
	IDxPstRelated  int    `db:"idx_pst_related"`
	IDxPstDateline int    `db:"idx_pst_dateline"`
	IDxPstContent  string `db:"idx_pst_content"`
}
