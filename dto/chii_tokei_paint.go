package dto

type ChiiTokeiPaint struct {
	TpID         int    `db:"tp_id"`
	TpUID        int    `db:"tp_uid"`
	TpHour       int    `db:"tp_hour"`
	TpMin        int    `db:"tp_min"`
	TpURL        string `db:"tp_url"`
	TpDesc       string `db:"tp_desc"`
	TpBook       int    `db:"tp_book"`
	TpViews      int    `db:"tp_views"`
	TpRelatedTpc int    `db:"tp_related_tpc"`
	TpLastupdate int    `db:"tp_lastupdate"`
	TpDateline   int    `db:"tp_dateline"`
}
