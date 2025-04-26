package dto

type ChiiPms struct {
	MsgID          int    `db:"msg_id"`
	MsgSID         int    `db:"msg_sid"`
	MsgRID         int    `db:"msg_rid"`
	MsgFolder      string `db:"msg_folder"`
	MsgNew         int    `db:"msg_new"`
	MsgTitle       string `db:"msg_title"`
	MsgDateline    int    `db:"msg_dateline"`
	MsgMessage     string `db:"msg_message"`
	MsgRelatedMain int    `db:"msg_related_main"`
	MsgRelated     int    `db:"msg_related"`
	MsgSdeleted    int    `db:"msg_sdeleted"`
	MsgRdeleted    int    `db:"msg_rdeleted"`
}
