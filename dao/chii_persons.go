package dto

type ChiiPersons struct {
	PrsnID          int    `db:"prsn_id"`
	PrsnName        string `db:"prsn_name"`
	PrsnType        int    `db:"prsn_type"`
	PrsnInfobox     string `db:"prsn_infobox"`
	PrsnProducer    int    `db:"prsn_producer"`
	PrsnMangaka     int    `db:"prsn_mangaka"`
	PrsnArtist      int    `db:"prsn_artist"`
	PrsnSeiyu       int    `db:"prsn_seiyu"`
	PrsnWriter      int    `db:"prsn_writer"`
	PrsnIllustrator int    `db:"prsn_illustrator"`
	PrsnActor       int    `db:"prsn_actor"`
	PrsnSummary     string `db:"prsn_summary"`
	PrsnImg         string `db:"prsn_img"`
	PrsnImgAnIDb    string `db:"prsn_img_anidb"`
	PrsnComment     int    `db:"prsn_comment"`
	PrsnCollects    int    `db:"prsn_collects"`
	PrsnDateline    int    `db:"prsn_dateline"`
	PrsnLastpost    int    `db:"prsn_lastpost"`
	PrsnLock        int    `db:"prsn_lock"`
	PrsnAnIDbId     int    `db:"prsn_anidb_id"`
	PrsnBan         int    `db:"prsn_ban"`
	PrsnRedirect    int    `db:"prsn_redirect"`
	PrsnNsfw        int    `db:"prsn_nsfw"`
}
