package dto

type ChiiDoujinPreorderReturn struct {
	RtID       int    `db:"rt_id"`
	RtPID      int    `db:"rt_pid"`
	RtUID      int    `db:"rt_uid"`
	RtStatus   int    `db:"rt_status"`
	RtJuiz     int    `db:"rt_juiz"`
	RtPaymail  string `db:"rt_paymail"`
	RtPhone    string `db:"rt_phone"`
	RtRealname string `db:"rt_realname"`
	RtUname    string `db:"rt_uname"`
	RtRemark   string `db:"rt_remark"`
	RtComment  string `db:"rt_comment"`
	RtDateline int    `db:"rt_dateline"`
}
