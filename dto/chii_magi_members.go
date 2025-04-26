package dto

type ChiiMagiMembers struct {
	MgmUID      int `db:"mgm_uid"`
	MgmCorrect  int `db:"mgm_correct"`
	MgmAnswered int `db:"mgm_answered"`
	MgmCreated  int `db:"mgm_created"`
	MgmRank     int `db:"mgm_rank"`
}
