package dto

import (
	"database/sql"
	"time"
)

type ChiiOauthAccessTokens struct {
	ID          int            `db:"id"`
	Type        int            `db:"type"`
	AccessToken string         `db:"access_token"`
	ClientID    string         `db:"client_id"`
	UserID      sql.NullString `db:"user_id"`
	Expires     time.Time      `db:"expires"`
	Scope       sql.NullString `db:"scope"`
	Info        string         `db:"info"`
}
