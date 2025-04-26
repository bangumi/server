package dto

type ChiiRobotSpeech struct {
	RbtSpcID       int    `db:"rbt_spc_id"`
	RbtSpcMID      int    `db:"rbt_spc_mid"`
	RbtSpcUID      int    `db:"rbt_spc_uid"`
	RbtSpcSpeech   string `db:"rbt_spc_speech"`
	RbtSpcBan      int    `db:"rbt_spc_ban"`
	RbtSpcDateline int    `db:"rbt_spc_dateline"`
}
