package dto

type ChiiSubjectRelations struct {
	RltSubjectID            int `db:"rlt_subject_id"`
	RltSubjectTypeID        int `db:"rlt_subject_type_id"`
	RltRelationType         int `db:"rlt_relation_type"`
	RltRelatedSubjectID     int `db:"rlt_related_subject_id"`
	RltRelatedSubjectTypeID int `db:"rlt_related_subject_type_id"`
	RltViceVersa            int `db:"rlt_vice_versa"`
	RltOrder                int `db:"rlt_order"`
}
