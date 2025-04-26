package dto

type ChiiSubjectTopics struct {
	SbjTpcID        int    `db:"sbj_tpc_id"`
	SbjTpcSubjectID int    `db:"sbj_tpc_subject_id"`
	SbjTpcUID       int    `db:"sbj_tpc_uid"`
	SbjTpcTitle     string `db:"sbj_tpc_title"`
	SbjTpcDateline  int    `db:"sbj_tpc_dateline"`
	SbjTpcLastpost  int    `db:"sbj_tpc_lastpost"`
	SbjTpcReplies   int    `db:"sbj_tpc_replies"`
	SbjTpcState     int    `db:"sbj_tpc_state"`
	SbjTpcDisplay   int    `db:"sbj_tpc_display"`
}
