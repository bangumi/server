package dto

type ChiiGroupPosts struct {
	GrpPstID       int    `db:"grp_pst_id"`
	GrpPstMID      int    `db:"grp_pst_mid"`
	GrpPstUID      int    `db:"grp_pst_uid"`
	GrpPstRelated  int    `db:"grp_pst_related"`
	GrpPstContent  string `db:"grp_pst_content"`
	GrpPstState    int    `db:"grp_pst_state"`
	GrpPstDateline int    `db:"grp_pst_dateline"`
}
