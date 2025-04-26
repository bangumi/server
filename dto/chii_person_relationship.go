package dto

type ChiiPersonRelationship struct {
	PrsnType      string `db:"prsn_type"`
	PrsnID        int    `db:"prsn_id"`
	RelatPrsnType string `db:"relat_prsn_type"`
	RelatPrsnID   int    `db:"relat_prsn_id"`
	RelatType     int    `db:"relat_type"`
}
