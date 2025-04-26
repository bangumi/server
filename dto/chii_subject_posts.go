package dto

type ChiiSubjectPosts struct {
	SbjPstID       int    `db:"sbj_pst_id"`
	SbjPstMID      int    `db:"sbj_pst_mid"`
	SbjPstUID      int    `db:"sbj_pst_uid"`
	SbjPstRelated  int    `db:"sbj_pst_related"`
	SbjPstContent  string `db:"sbj_pst_content"`
	SbjPstState    int    `db:"sbj_pst_state"`
	SbjPstDateline int    `db:"sbj_pst_dateline"`
}
