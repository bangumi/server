package dto

type ChiiEventPosts struct {
	EventPstID       int    `db:"event_pst_id"`
	EventPstMID      int    `db:"event_pst_mid"`
	EventPstUID      int    `db:"event_pst_uid"`
	EventPstRelated  int    `db:"event_pst_related"`
	EventPstContent  string `db:"event_pst_content"`
	EventPstDateline int    `db:"event_pst_dateline"`
}
