package dto

type ChiiBlogComments struct {
	BlgPstID       int    `db:"blg_pst_id"`
	BlgPstMID      int    `db:"blg_pst_mid"`
	BlgPstUID      int    `db:"blg_pst_uid"`
	BlgPstRelated  int    `db:"blg_pst_related"`
	BlgPstDateline int    `db:"blg_pst_dateline"`
	BlgPstContent  string `db:"blg_pst_content"`
}
