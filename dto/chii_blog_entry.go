package dto

type ChiiBlogEntry struct {
	EntryID       int    `db:"entry_id"`
	EntryType     int    `db:"entry_type"`
	EntryUID      int    `db:"entry_uid"`
	EntryTitle    string `db:"entry_title"`
	EntryIcon     string `db:"entry_icon"`
	EntryContent  string `db:"entry_content"`
	EntryTags     string `db:"entry_tags"`
	EntryViews    int    `db:"entry_views"`
	EntryReplies  int    `db:"entry_replies"`
	EntryDateline int    `db:"entry_dateline"`
	EntryLastpost int    `db:"entry_lastpost"`
	EntryLike     int    `db:"entry_like"`
	EntryDislike  int    `db:"entry_dislike"`
	EntryNoreply  int    `db:"entry_noreply"`
	EntryRelated  int    `db:"entry_related"`
	EntryPublic   int    `db:"entry_public"`
}
