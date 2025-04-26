package dto

type ChiiPersonCollects struct {
	PrsnCltID       int    `db:"prsn_clt_id"`
	PrsnCltCat      string `db:"prsn_clt_cat"`
	PrsnCltMID      int    `db:"prsn_clt_mid"`
	PrsnCltUID      int    `db:"prsn_clt_uid"`
	PrsnCltDateline int    `db:"prsn_clt_dateline"`
}
