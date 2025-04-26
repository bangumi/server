package dto

type ChiiPrsnComments struct {
	PrsnPstID       int    `db:"prsn_pst_id"`
	PrsnPstMID      int    `db:"prsn_pst_mid"`
	PrsnPstUID      int    `db:"prsn_pst_uid"`
	PrsnPstRelated  int    `db:"prsn_pst_related"`
	PrsnPstDateline int    `db:"prsn_pst_dateline"`
	PrsnPstContent  string `db:"prsn_pst_content"`
	PrsnPstState    int    `db:"prsn_pst_state"`
}
