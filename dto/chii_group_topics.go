package dto

type ChiiGroupTopics struct {
	GrpTpcID       int    `db:"grp_tpc_id"`
	GrpTpcGID      int    `db:"grp_tpc_gid"`
	GrpTpcUID      int    `db:"grp_tpc_uid"`
	GrpTpcTitle    string `db:"grp_tpc_title"`
	GrpTpcDateline int    `db:"grp_tpc_dateline"`
	GrpTpcLastpost int    `db:"grp_tpc_lastpost"`
	GrpTpcReplies  int    `db:"grp_tpc_replies"`
	GrpTpcState    int    `db:"grp_tpc_state"`
	GrpTpcDisplay  int    `db:"grp_tpc_display"`
}
