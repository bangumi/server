package dto

type ChiiDoujinPreorder struct {
	PreID       int    `db:"pre_id"`
	PreUID      int    `db:"pre_uid"`
	PreType     int    `db:"pre_type"`
	PreMID      int    `db:"pre_mid"`
	PreDetails  string `db:"pre_details"`
	PreDateline int    `db:"pre_dateline"`
}
