package dto

import (
	"database/sql"
)

type ChiiOauthJwt struct {
	ClientID  string         `db:"client_id"`
	Subject   sql.NullString `db:"subject"`
	PublicKey string         `db:"public_key"`
}
