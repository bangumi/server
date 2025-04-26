package dto

import (
	"database/sql"
	"time"
)

type ChiiOauthAuthorizationCodes struct {
	AuthorizationCode string         `db:"authorization_code"`
	ClientID          string         `db:"client_id"`
	UserID            sql.NullString `db:"user_id"`
	RedirectUri       sql.NullString `db:"redirect_uri"`
	Expires           time.Time      `db:"expires"`
	Scope             sql.NullString `db:"scope"`
	IDToken           sql.NullString `db:"id_token"`
}
