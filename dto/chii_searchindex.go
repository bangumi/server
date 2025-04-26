package dto

type ChiiSearchindex struct {
	Keywords     string `db:"keywords"`
	Searchstring string `db:"searchstring"`
	Dateline     int    `db:"dateline"`
	Expiration   int    `db:"expiration"`
	Threads      int    `db:"threads"`
	TIDs         string `db:"tids"`
	Type         string `db:"type"`
}
