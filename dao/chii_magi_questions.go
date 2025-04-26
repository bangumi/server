package dto

type ChiiMagiQuestions struct {
	QstID         int    `db:"qst_id"`
	QstType       int    `db:"qst_type"`
	QstContent    string `db:"qst_content"`
	QstOptions    string `db:"qst_options"`
	QstAnswer     int    `db:"qst_answer"`
	QstRelateType int    `db:"qst_relate_type"`
	QstRelated    int    `db:"qst_related"`
	QstCorrect    int    `db:"qst_correct"`
	QstAnswered   int    `db:"qst_answered"`
	QstCreator    int    `db:"qst_creator"`
	QstDateline   int    `db:"qst_dateline"`
	QstBan        int    `db:"qst_ban"`
}
