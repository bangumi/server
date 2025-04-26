package dto

import (
	"database/sql"
)

type ChiiOauthClients struct {
	AppID        int            `db:"app_id"`
	ClientID     string         `db:"client_id"`
	ClientSecret sql.NullString `db:"client_secret"`
	RedirectUri  sql.NullString `db:"redirect_uri"`
	GrantTypes   sql.NullString `db:"grant_types"`
	Scope        sql.NullString `db:"scope"`
	UserID       sql.NullString `db:"user_id"`
}
