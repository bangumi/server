package model

import "time"

type IndexComment struct {
	ID        CommentID `json:"id"`
	Field     IndexID   `json:"field"`
	User      UserID    `json:"user"`
	Related   CommentID `json:"related"` // 回复消息的ID
	CreatedAt time.Time `json:"created_at"`
	Content   string    `json:"content"`
}

type SubjectPost struct {
	ID        CommentID     `json:"id"`
	Field     SubjectID     `json:"field"`
	User      UserID        `json:"user"`
	Related   CommentID     `json:"related"`
	CreatedAt time.Time     `json:"created_at"`
	Content   string        `json:"content"`
	State     uint8         `json:"state"`
	Replies   []SubjectPost `json:"replies"`
}

type EpisodeComment struct {
	ID        CommentID `json:"id"`
	Field     EpisodeID `json:"field"`
	User      UserID    `json:"user"`
	Related   CommentID `json:"related"` // 回复消息的ID
	CreatedAt time.Time `json:"created_at"`
	Content   string    `json:"content"`
}
