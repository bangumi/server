package dto

type ChiiSubjectImgs struct {
	ImgID        int    `db:"img_id"`
	ImgSubjectID int    `db:"img_subject_id"`
	ImgUID       int    `db:"img_uid"`
	ImgTarget    string `db:"img_target"`
	ImgVote      int    `db:"img_vote"`
	ImgNsfw      int    `db:"img_nsfw"`
	ImgBan       int    `db:"img_ban"`
	ImgDateline  int    `db:"img_dateline"`
}
