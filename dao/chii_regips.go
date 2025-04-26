package dto

type ChiiRegips struct {
	Ip       string `db:"ip"`
	Dateline int    `db:"dateline"`
	Count    int    `db:"count"`
}
