package dto

type ChiiDoujinInvites struct {
	UID        int    `db:"uid"`
	Dateline   int    `db:"dateline"`
	Invitecode string `db:"invitecode"`
	Status     int    `db:"status"`
	InviteUID  int    `db:"invite_uid"`
}
