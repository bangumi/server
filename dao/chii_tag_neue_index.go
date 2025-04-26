package dto

type ChiiTagNeueIndex struct {
	TagID        int    `db:"tag_id"`
	TagName      string `db:"tag_name"`
	TagCat       int    `db:"tag_cat"`
	TagType      int    `db:"tag_type"`
	TagResults   int    `db:"tag_results"`
	TagDateline  int    `db:"tag_dateline"`
	TagLasttouch int    `db:"tag_lasttouch"`
}
