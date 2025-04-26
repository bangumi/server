package dto

type ChiiDoujinSubjects struct {
	SubjectID        int    `db:"subject_id"`
	SubjectType      int    `db:"subject_type"`
	SubjectCat       int    `db:"subject_cat"`
	SubjectName      string `db:"subject_name"`
	SubjectInfobox   string `db:"subject_infobox"`
	SubjectDesc      string `db:"subject_desc"`
	SubjectImg       string `db:"subject_img"`
	SubjectCollects  int    `db:"subject_collects"`
	SubjectStatus    int    `db:"subject_status"`
	SubjectOriginal  int    `db:"subject_original"`
	SubjectSexual    int    `db:"subject_sexual"`
	SubjectAgeLimit  int    `db:"subject_age_limit"`
	SubjectTags      string `db:"subject_tags"`
	SubjectAttrTags  string `db:"subject_attr_tags"`
	SubjectCreator   int    `db:"subject_creator"`
	SubjectComment   int    `db:"subject_comment"`
	SubjectDateline  int    `db:"subject_dateline"`
	SubjectLasttouch int    `db:"subject_lasttouch"`
	SubjectLastpost  int    `db:"subject_lastpost"`
	SubjectBan       int    `db:"subject_ban"`
	SubjectBanReason string `db:"subject_ban_reason"`
}
