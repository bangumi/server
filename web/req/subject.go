package req

import "github.com/bangumi/server/internal/model"

type SubjectComment struct {
	ID      model.CommentID `json:"id,,omitempty"`
	FieldID model.CommentID `json:"field_id,omitempty"`
	Content string          `json:"content"`
}
