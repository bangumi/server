package dto

type ChiiSubjectInterests struct {
	InterestID              int    `db:"interest_id"`
	InterestUID             int    `db:"interest_uid"`
	InterestSubjectID       int    `db:"interest_subject_id"`
	InterestSubjectType     int    `db:"interest_subject_type"`
	InterestRate            int    `db:"interest_rate"`
	InterestType            int    `db:"interest_type"`
	InterestHasComment      int    `db:"interest_has_comment"`
	InterestComment         string `db:"interest_comment"`
	InterestTag             string `db:"interest_tag"`
	InterestEpStatus        int    `db:"interest_ep_status"`
	InterestVolStatus       int    `db:"interest_vol_status"`
	InterestWishDateline    int    `db:"interest_wish_dateline"`
	InterestDoingDateline   int    `db:"interest_doing_dateline"`
	InterestCollectDateline int    `db:"interest_collect_dateline"`
	InterestOnHoldDateline  int    `db:"interest_on_hold_dateline"`
	InterestDroppedDateline int    `db:"interest_dropped_dateline"`
	InterestCreateIp        string `db:"interest_create_ip"`
	InterestLasttouchIp     string `db:"interest_lasttouch_ip"`
	InterestLasttouch       int    `db:"interest_lasttouch"`
	InterestPrivate         int    `db:"interest_private"`
}
