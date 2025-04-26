package dto

type ChiiLikesBak230331 struct {
	Type      int `db:"type"`
	MainID    int `db:"main_id"`
	RelatedID int `db:"related_id"`
	UID       int `db:"uid"`
	Value     int `db:"value"`
	Ban       int `db:"ban"`
	CreatedAt int `db:"created_at"`
}
