package dto

type ChiiDoujinClubs struct {
	ClubID         int    `db:"club_id"`
	ClubType       int    `db:"club_type"`
	ClubName       string `db:"club_name"`
	ClubTitle      string `db:"club_title"`
	ClubIcon       string `db:"club_icon"`
	ClubCreator    int    `db:"club_creator"`
	ClubProBook    int    `db:"club_pro_book"`
	ClubProMusic   int    `db:"club_pro_music"`
	ClubProGame    int    `db:"club_pro_game"`
	ClubMembers    int    `db:"club_members"`
	ClubFollowers  int    `db:"club_followers"`
	ClubDesc       string `db:"club_desc"`
	ClubBuilddate  int    `db:"club_builddate"`
	ClubLastupdate int    `db:"club_lastupdate"`
	ClubBan        int    `db:"club_ban"`
}
