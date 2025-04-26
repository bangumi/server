package dto

type ChiiDoujinClubTopics struct {
	ClubTpcID       int    `db:"club_tpc_id"`
	ClubTpcClubID   int    `db:"club_tpc_club_id"`
	ClubTpcUID      int    `db:"club_tpc_uid"`
	ClubTpcTitle    string `db:"club_tpc_title"`
	ClubTpcDateline int    `db:"club_tpc_dateline"`
	ClubTpcLastpost int    `db:"club_tpc_lastpost"`
	ClubTpcReplies  int    `db:"club_tpc_replies"`
	ClubTpcDisplay  int    `db:"club_tpc_display"`
}
