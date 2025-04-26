package dto

type ChiiDoujinClubFields struct {
	CfCID    int    `db:"cf_cid"`
	CfHeader string `db:"cf_header"`
	CfBg     string `db:"cf_bg"`
	CfTheme  int    `db:"cf_theme"`
	CfDesign string `db:"cf_design"`
	CfModel  string `db:"cf_model"`
}
