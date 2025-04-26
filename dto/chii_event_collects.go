package dto

type ChiiEventCollects struct {
	CltUID      int `db:"clt_uid"`
	CltEventID  int `db:"clt_event_id"`
	CltType     int `db:"clt_type"`
	CltDateline int `db:"clt_dateline"`
}
