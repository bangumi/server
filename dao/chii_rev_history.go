package dto

type ChiiRevHistory struct {
	RevID          int    `db:"rev_id"`
	RevType        int    `db:"rev_type"`
	RevMID         int    `db:"rev_mid"`
	RevTextID      int    `db:"rev_text_id"`
	RevDateline    int    `db:"rev_dateline"`
	RevCreator     int    `db:"rev_creator"`
	RevEditSummary string `db:"rev_edit_summary"`
}
