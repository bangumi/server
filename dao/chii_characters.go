package dto

type ChiiCharacters struct {
	CrtID       int    `db:"crt_id"`
	CrtName     string `db:"crt_name"`
	CrtRole     int    `db:"crt_role"`
	CrtInfobox  string `db:"crt_infobox"`
	CrtSummary  string `db:"crt_summary"`
	CrtImg      string `db:"crt_img"`
	CrtComment  int    `db:"crt_comment"`
	CrtCollects int    `db:"crt_collects"`
	CrtDateline int    `db:"crt_dateline"`
	CrtLastpost int    `db:"crt_lastpost"`
	CrtLock     int    `db:"crt_lock"`
	CrtImgAnIDb string `db:"crt_img_anidb"`
	CrtAnIDbId  int    `db:"crt_anidb_id"`
	CrtBan      int    `db:"crt_ban"`
	CrtRedirect int    `db:"crt_redirect"`
	CrtNsfw     int    `db:"crt_nsfw"`
}
