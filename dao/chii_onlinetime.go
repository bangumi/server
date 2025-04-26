package dto

type ChiiOnlinetime struct {
	UID        int `db:"uid"`
	Thismonth  int `db:"thismonth"`
	Total      int `db:"total"`
	Lastupdate int `db:"lastupdate"`
}
