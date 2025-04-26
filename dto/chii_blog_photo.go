package dto

type ChiiBlogPhoto struct {
	PhotoID       int    `db:"photo_id"`
	PhotoEID      int    `db:"photo_eid"`
	PhotoUID      int    `db:"photo_uid"`
	PhotoTarget   string `db:"photo_target"`
	PhotoVote     int    `db:"photo_vote"`
	PhotoDateline int    `db:"photo_dateline"`
}
