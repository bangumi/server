package dto

type ChiiDoujinSubjectPhotos struct {
	SbjPhotoID        int    `db:"sbj_photo_id"`
	SbjPhotoMID       int    `db:"sbj_photo_mid"`
	SbjPhotoUID       int    `db:"sbj_photo_uid"`
	SbjPhotoTarget    string `db:"sbj_photo_target"`
	SbjPhotoComment   int    `db:"sbj_photo_comment"`
	SbjPhotoDateline  int    `db:"sbj_photo_dateline"`
	SbjPhotoLasttouch int    `db:"sbj_photo_lasttouch"`
	SbjPhotoBan       int    `db:"sbj_photo_ban"`
}
