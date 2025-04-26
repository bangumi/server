package dto

type ChiiUsergroup struct {
	UsrGrpID       int    `db:"usr_grp_id"`
	UsrGrpName     string `db:"usr_grp_name"`
	UsrGrpPerm     string `db:"usr_grp_perm"`
	UsrGrpDateline int    `db:"usr_grp_dateline"`
}
