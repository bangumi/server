package dto

type ChiiGroupMembers struct {
	GmbUID       int `db:"gmb_uid"`
	GmbGID       int `db:"gmb_gid"`
	GmbModerator int `db:"gmb_moderator"`
	GmbBan       int `db:"gmb_ban"`
	GmbDateline  int `db:"gmb_dateline"`
}
