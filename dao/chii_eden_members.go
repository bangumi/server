package dto

type ChiiEdenMembers struct {
	EmbUID       int `db:"emb_uid"`
	EmbEID       int `db:"emb_eid"`
	EmbModerator int `db:"emb_moderator"`
	EmbDateline  int `db:"emb_dateline"`
}
