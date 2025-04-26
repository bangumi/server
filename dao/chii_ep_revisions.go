package dto

type ChiiEpRevisions struct {
	EpRevID        int    `db:"ep_rev_id"`
	RevSID         int    `db:"rev_sid"`
	RevEIDs        string `db:"rev_eids"`
	RevEpInfobox   string `db:"rev_ep_infobox"`
	RevCreator     int    `db:"rev_creator"`
	RevVersion     int    `db:"rev_version"`
	RevDateline    int    `db:"rev_dateline"`
	RevEditSummary string `db:"rev_edit_summary"`
}
