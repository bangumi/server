package dto

type ChiiEventTopics struct {
	EventTpcID       int    `db:"event_tpc_id"`
	EventTpcMID      int    `db:"event_tpc_mid"`
	EventTpcUID      int    `db:"event_tpc_uid"`
	EventTpcTitle    string `db:"event_tpc_title"`
	EventTpcDateline int    `db:"event_tpc_dateline"`
	EventTpcLastpost int    `db:"event_tpc_lastpost"`
	EventTpcReplies  int    `db:"event_tpc_replies"`
	EventTpcDisplay  int    `db:"event_tpc_display"`
}
