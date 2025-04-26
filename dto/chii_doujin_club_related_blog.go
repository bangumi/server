package dto

type ChiiDoujinClubRelatedBlog struct {
	CrbID       int `db:"crb_id"`
	CrbUID      int `db:"crb_uid"`
	CrbClubID   int `db:"crb_club_id"`
	CrbEntryID  int `db:"crb_entry_id"`
	CrbStick    int `db:"crb_stick"`
	CrbDateline int `db:"crb_dateline"`
}
