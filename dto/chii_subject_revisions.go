package dto

type ChiiSubjectRevisions struct {
	RevID            int    `db:"rev_id"`
	RevType          int    `db:"rev_type"`
	RevSubjectID     int    `db:"rev_subject_id"`
	RevTypeID        int    `db:"rev_type_id"`
	RevCreator       int    `db:"rev_creator"`
	RevDateline      int    `db:"rev_dateline"`
	RevName          string `db:"rev_name"`
	RevNameCn        string `db:"rev_name_cn"`
	RevFieldInfobox  string `db:"rev_field_infobox"`
	RevFieldMetaTags string `db:"rev_field_meta_tags"`
	RevFieldSummary  string `db:"rev_field_summary"`
	RevVoteField     string `db:"rev_vote_field"`
	RevFieldEps      int    `db:"rev_field_eps"`
	RevEditSummary   string `db:"rev_edit_summary"`
	RevPlatform      int    `db:"rev_platform"`
}
