package dto

type ChiiSubjects struct {
	SubjectID          int    `db:"subject_id"`
	SubjectTypeID      int    `db:"subject_type_id"`
	SubjectName        string `db:"subject_name"`
	SubjectNameCn      string `db:"subject_name_cn"`
	SubjectUID         string `db:"subject_uid"`
	SubjectCreator     int    `db:"subject_creator"`
	SubjectDateline    int    `db:"subject_dateline"`
	SubjectImage       string `db:"subject_image"`
	SubjectPlatform    int    `db:"subject_platform"`
	FieldInfobox       string `db:"field_infobox"`
	FieldMetaTags      string `db:"field_meta_tags"`
	FieldSummary       string `db:"field_summary"`
	Field5             string `db:"field_5"`
	FieldVolumes       int    `db:"field_volumes"`
	FieldEps           int    `db:"field_eps"`
	SubjectWish        int    `db:"subject_wish"`
	SubjectCollect     int    `db:"subject_collect"`
	SubjectDoing       int    `db:"subject_doing"`
	SubjectOnHold      int    `db:"subject_on_hold"`
	SubjectDropped     int    `db:"subject_dropped"`
	SubjectSeries      int    `db:"subject_series"`
	SubjectSeriesEntry int    `db:"subject_series_entry"`
	SubjectIDxCn       string `db:"subject_idx_cn"`
	SubjectAirtime     int    `db:"subject_airtime"`
	SubjectNsfw        int    `db:"subject_nsfw"`
	SubjectBan         int    `db:"subject_ban"`
}
