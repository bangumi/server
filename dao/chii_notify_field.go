package dto

type ChiiNotifyField struct {
	NtfID    int    `db:"ntf_id"`
	NtfHash  int    `db:"ntf_hash"`
	NtfRID   int    `db:"ntf_rid"`
	NtfTitle string `db:"ntf_title"`
}
