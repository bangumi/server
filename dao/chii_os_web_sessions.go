package dto

type ChiiOsWebSessions struct {
	Key       string `db:"key"`
	UserID    int    `db:"user_id"`
	Value     string `db:"value"`
	CreatedAt int    `db:"created_at"`
	ExpiredAt int    `db:"expired_at"`
}
