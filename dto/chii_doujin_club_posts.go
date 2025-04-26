package dto

type ChiiDoujinClubPosts struct {
	ClubPstID       int    `db:"club_pst_id"`
	ClubPstMID      int    `db:"club_pst_mid"`
	ClubPstUID      int    `db:"club_pst_uid"`
	ClubPstRelated  int    `db:"club_pst_related"`
	ClubPstContent  string `db:"club_pst_content"`
	ClubPstDateline int    `db:"club_pst_dateline"`
}
