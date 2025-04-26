package dto

type ChiiCrtRevisions struct {
	RevID          int    `db:"rev_id"`
	RevCrtID       int    `db:"rev_crt_id"`
	RevCrtName     string `db:"rev_crt_name"`
	RevCrtNameCn   string `db:"rev_crt_name_cn"`
	RevCrtInfoWiki string `db:"rev_crt_info_wiki"`
	RevCrtSummary  string `db:"rev_crt_summary"`
	RevDateline    int    `db:"rev_dateline"`
	RevCreator     int    `db:"rev_creator"`
	RevEditSummary string `db:"rev_edit_summary"`
}
