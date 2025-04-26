package dto

import (
	"database/sql"
)

type ChiiOauthScopes struct {
	Scope     string        `db:"scope"`
	IsDefault sql.NullInt64 `db:"is_default"`
}
