package dto

type ChiiTimelineComments struct {
	TmlPstID       int    `db:"tml_pst_id"`
	TmlPstMID      int    `db:"tml_pst_mid"`
	TmlPstUID      int    `db:"tml_pst_uid"`
	TmlPstRelated  int    `db:"tml_pst_related"`
	TmlPstDateline int    `db:"tml_pst_dateline"`
	TmlPstContent  string `db:"tml_pst_content"`
}
