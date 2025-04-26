package dto

type ChiiRobotPersonality struct {
	RbtPsnID        int    `db:"rbt_psn_id"`
	RbtPsnName      string `db:"rbt_psn_name"`
	RbtPsnCreator   int    `db:"rbt_psn_creator"`
	RbtPsnDesc      string `db:"rbt_psn_desc"`
	RbtPsnSpeech    int    `db:"rbt_psn_speech"`
	RbtPsnBan       int    `db:"rbt_psn_ban"`
	RbtPsnLasttouch int    `db:"rbt_psn_lasttouch"`
	RbtPsnDateline  int    `db:"rbt_psn_dateline"`
}
