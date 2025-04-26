package dto

type ChiiFriends struct {
	FrdUID         int    `db:"frd_uid"`
	FrdFID         int    `db:"frd_fid"`
	FrdGrade       int    `db:"frd_grade"`
	FrdDateline    int    `db:"frd_dateline"`
	FrdDescription string `db:"frd_description"`
}
