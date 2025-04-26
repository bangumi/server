package dto

type ChiiDoujinSubjectComments struct {
	SbjPstID           int    `db:"sbj_pst_id"`
	SbjPstMID          int    `db:"sbj_pst_mid"`
	SbjPstUID          int    `db:"sbj_pst_uid"`
	SbjPstRelated      int    `db:"sbj_pst_related"`
	SbjPstRelatedPhoto int    `db:"sbj_pst_related_photo"`
	SbjPstDateline     int    `db:"sbj_pst_dateline"`
	SbjPstContent      string `db:"sbj_pst_content"`
}
