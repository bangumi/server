package dto

type ChiiFailedlogins struct {
	Ip         string `db:"ip"`
	Count      int    `db:"count"`
	Lastupdate int    `db:"lastupdate"`
}
