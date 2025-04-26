package dto

type ChiiRevText struct {
	RevTextID int    `db:"rev_text_id"`
	RevText   string `db:"rev_text"`
}
