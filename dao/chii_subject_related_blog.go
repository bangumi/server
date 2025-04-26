package dto

type ChiiSubjectRelatedBlog struct {
	SrbID        int `db:"srb_id"`
	SrbUID       int `db:"srb_uid"`
	SrbSubjectID int `db:"srb_subject_id"`
	SrbEntryID   int `db:"srb_entry_id"`
	SrbSpoiler   int `db:"srb_spoiler"`
	SrbLike      int `db:"srb_like"`
	SrbDislike   int `db:"srb_dislike"`
	SrbDateline  int `db:"srb_dateline"`
}
