package dto

type ChiiNetworkServices struct {
	NsUID       int    `db:"ns_uid"`
	NsServiceID int    `db:"ns_service_id"`
	NsAccount   string `db:"ns_account"`
	NsDateline  int    `db:"ns_dateline"`
}
