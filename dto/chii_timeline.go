package dto

type ChiiTimeline struct {
	TmlID       int    `db:"tml_id"`
	TmlUID      int    `db:"tml_uid"`
	TmlCat      int    `db:"tml_cat"`
	TmlType     int    `db:"tml_type"`
	TmlRelated  string `db:"tml_related"`
	TmlMemo     string `db:"tml_memo"`
	TmlImg      string `db:"tml_img"`
	TmlBatch    int    `db:"tml_batch"`
	TmlSource   int    `db:"tml_source"`
	TmlReplies  int    `db:"tml_replies"`
	TmlDateline int    `db:"tml_dateline"`
}
