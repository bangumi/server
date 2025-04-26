package dto

type ChiiSubjectAlias struct {
	SubjectID     int    `db:"subject_id"`
	AliasName     string `db:"alias_name"`
	SubjectTypeID int    `db:"subject_type_id"`
	AliasType     int    `db:"alias_type"`
	AliasKey      string `db:"alias_key"`
}
