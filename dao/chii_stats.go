package dto

type ChiiStats struct {
	Unit      int `db:"unit"`
	Category  int `db:"category"`
	Type      int `db:"type"`
	SubType   int `db:"sub_type"`
	RelatedID int `db:"related_id"`
	Value     int `db:"value"`
	Timestamp int `db:"timestamp"`
	UpdatedAt int `db:"updated_at"`
}
