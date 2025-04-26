package dto

type ChiiEpComments struct {
	EpPstID       int    `db:"ep_pst_id"`
	EpPstMID      int    `db:"ep_pst_mid"`
	EpPstUID      int    `db:"ep_pst_uid"`
	EpPstRelated  int    `db:"ep_pst_related"`
	EpPstDateline int    `db:"ep_pst_dateline"`
	EpPstContent  string `db:"ep_pst_content"`
	EpPstState    int    `db:"ep_pst_state"`
}
