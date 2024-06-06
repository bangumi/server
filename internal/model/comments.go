package model

import "time"

type IndexComment struct {
	ID        CommentID
	Field     IndexID
	User      UserID
	Related   CommentID // 回复消息的ID
	CreatedAt time.Time
	Content   string
}

type SubjectPost struct {
	ID        CommentID
	Field     SubjectID
	User      UserID
	Related   CommentID
	CreatedAt time.Time
	Content   string
	State     uint8
}

type EpisodeComment struct {
	ID        CommentID
	Field     EpisodeID
	User      UserID
	Related   CommentID // 回复消息的ID
	CreatedAt time.Time
	Content   string
}
