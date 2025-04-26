package dto

import (
	"time"
)

type ChiiPersonFields struct {
	PrsnCat   string    `db:"prsn_cat"`
	PrsnID    int       `db:"prsn_id"`
	Gender    int       `db:"gender"`
	Bloodtype int       `db:"bloodtype"`
	BirthYear time.Time `db:"birth_year"`
	BirthMon  int       `db:"birth_mon"`
	BirthDay  int       `db:"birth_day"`
}
