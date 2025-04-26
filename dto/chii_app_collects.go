package dto

type ChiiAppCollects struct {
	AppCltID       int `db:"app_clt_id"`
	AppCltAppID    int `db:"app_clt_app_id"`
	AppCltUID      int `db:"app_clt_uid"`
	AppCltDateline int `db:"app_clt_dateline"`
}
