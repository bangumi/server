package dto

type ChiiPersonAlias struct {
	PrsnCat   string `db:"prsn_cat"`
	PrsnID    int    `db:"prsn_id"`
	AliasName string `db:"alias_name"`
	AliasType int    `db:"alias_type"`
	AliasKey  string `db:"alias_key"`
}
