package dto

import (
	"time"
)

type ChiiSubjectFields struct {
	FieldSID      int       `db:"field_sid"`
	FieldTID      int       `db:"field_tid"`
	FieldTags     string    `db:"field_tags"`
	FieldRate1    int       `db:"field_rate_1"`
	FieldRate2    int       `db:"field_rate_2"`
	FieldRate3    int       `db:"field_rate_3"`
	FieldRate4    int       `db:"field_rate_4"`
	FieldRate5    int       `db:"field_rate_5"`
	FieldRate6    int       `db:"field_rate_6"`
	FieldRate7    int       `db:"field_rate_7"`
	FieldRate8    int       `db:"field_rate_8"`
	FieldRate9    int       `db:"field_rate_9"`
	FieldRate10   int       `db:"field_rate_10"`
	FieldAirtime  int       `db:"field_airtime"`
	FieldRank     int       `db:"field_rank"`
	FieldYear     time.Time `db:"field_year"`
	FieldMon      int       `db:"field_mon"`
	FieldWeekDay  int       `db:"field_week_day"`
	FieldDate     time.Time `db:"field_date"`
	FieldRedirect int       `db:"field_redirect"`
}
