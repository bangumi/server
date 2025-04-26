package dto

type ChiiDoujinClubTimeline struct {
	TmlID       int    `db:"tml_id"`
	TmlCID      int    `db:"tml_cid"`
	TmlType     int    `db:"tml_type"`
	TmlRelated  string `db:"tml_related"`
	TmlMemo     string `db:"tml_memo"`
	TmlDateline int    `db:"tml_dateline"`
}
