package dto

import (
	"database/sql"
	"time"
)

type ChiiOauthRefreshTokens struct {
	RefreshToken string         `db:"refresh_token"`
	ClientID     string         `db:"client_id"`
	UserID       sql.NullString `db:"user_id"`
	Expires      time.Time      `db:"expires"`
	Scope        sql.NullString `db:"scope"`
}
