package dto

type ChiiDoujinClubMembers struct {
	CmbUID       int    `db:"cmb_uid"`
	CmbCID       int    `db:"cmb_cid"`
	CmbModerator int    `db:"cmb_moderator"`
	CmbPerm      string `db:"cmb_perm"`
	CmbNote      string `db:"cmb_note"`
	CmbDeteline  int    `db:"cmb_deteline"`
}
