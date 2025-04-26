package dto

type ChiiDoujinSubjectClubIndex struct {
	SubjectID   int `db:"subject_id"`
	ClubID      int `db:"club_id"`
	SubjectType int `db:"subject_type"`
	ClubRole    int `db:"club_role"`
}
