package dto

type ChiiCrtComments struct {
	CrtPstID       int    `db:"crt_pst_id"`
	CrtPstMID      int    `db:"crt_pst_mid"`
	CrtPstUID      int    `db:"crt_pst_uid"`
	CrtPstRelated  int    `db:"crt_pst_related"`
	CrtPstDateline int    `db:"crt_pst_dateline"`
	CrtPstContent  string `db:"crt_pst_content"`
	CrtPstState    int    `db:"crt_pst_state"`
}
