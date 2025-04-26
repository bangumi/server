package dto

type ChiiGadgets struct {
	GdtID          int    `db:"gdt_id"`
	GdtAppID       int    `db:"gdt_app_id"`
	GdtVersion     string `db:"gdt_version"`
	GdtCreator     int    `db:"gdt_creator"`
	GdtScript      string `db:"gdt_script"`
	GdtStyle       string `db:"gdt_style"`
	GdtHasScript   int    `db:"gdt_has_script"`
	GdtHasStyle    int    `db:"gdt_has_style"`
	GdtStatus      int    `db:"gdt_status"`
	GdtEditSummary string `db:"gdt_edit_summary"`
	GdtTimestamp   int    `db:"gdt_timestamp"`
	GdtLasttouch   int    `db:"gdt_lasttouch"`
}
