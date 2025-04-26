package dto

type ChiiSessions struct {
	SID          string `db:"sid"`
	Ip1          int    `db:"ip1"`
	Ip2          int    `db:"ip2"`
	Ip3          int    `db:"ip3"`
	Ip4          int    `db:"ip4"`
	UID          int    `db:"uid"`
	Username     string `db:"username"`
	GroupID      int    `db:"groupid"`
	StyleID      int    `db:"styleid"`
	Invisible    int    `db:"invisible"`
	Action       int    `db:"action"`
	Lastactivity int    `db:"lastactivity"`
	Lastolupdate int    `db:"lastolupdate"`
	Pageviews    int    `db:"pageviews"`
	Seccode      int    `db:"seccode"`
	FID          int    `db:"fid"`
	TID          int    `db:"tid"`
	BloguID      int    `db:"bloguid"`
}
