package dto

type ChiiTagNeueFields struct {
	FieldTID     int    `db:"field_tid"`
	FieldSummary string `db:"field_summary"`
	FieldOrder   int    `db:"field_order"`
	FieldNsfw    int    `db:"field_nsfw"`
	FieldLock    int    `db:"field_lock"`
}
