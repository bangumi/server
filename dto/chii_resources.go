package dto

type ChiiResources struct {
	ResID        int    `db:"res_id"`
	ResEID       int    `db:"res_eid"`
	ResType      int    `db:"res_type"`
	ResTool      int    `db:"res_tool"`
	ResURL       string `db:"res_url"`
	ResExt       string `db:"res_ext"`
	ResAudioLang int    `db:"res_audio_lang"`
	ResSubLang   int    `db:"res_sub_lang"`
	ResQuality   int    `db:"res_quality"`
	ResSource    string `db:"res_source"`
	ResVersion   int    `db:"res_version"`
	ResCreator   int    `db:"res_creator"`
	ResDateline  int    `db:"res_dateline"`
}
