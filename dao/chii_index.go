package dto

type ChiiIndex struct {
	IDxID           int    `db:"idx_id"`
	IDxType         int    `db:"idx_type"`
	IDxTitle        string `db:"idx_title"`
	IDxDesc         string `db:"idx_desc"`
	IDxReplies      int    `db:"idx_replies"`
	IDxSubjectTotal int    `db:"idx_subject_total"`
	IDxCollects     int    `db:"idx_collects"`
	IDxStats        string `db:"idx_stats"`
	IDxAward        int    `db:"idx_award"`
	IDxDateline     int    `db:"idx_dateline"`
	IDxLasttouch    int    `db:"idx_lasttouch"`
	IDxUid          int    `db:"idx_uid"`
	IDxBan          int    `db:"idx_ban"`
}
