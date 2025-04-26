package dto

type ChiiEventClubIndex struct {
	EventID   int    `db:"event_id"`
	ClubID    int    `db:"club_id"`
	ClubPlace string `db:"club_place"`
	Dateline  int    `db:"dateline"`
}
