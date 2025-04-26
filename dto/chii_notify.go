package dto

type ChiiNotify struct {
	NtID        int `db:"nt_id"`
	NtUID       int `db:"nt_uid"`
	NtFromUID   int `db:"nt_from_uid"`
	NtStatus    int `db:"nt_status"`
	NtType      int `db:"nt_type"`
	NtMID       int `db:"nt_mid"`
	NtRelatedID int `db:"nt_related_id"`
	NtDateline  int `db:"nt_dateline"`
}
