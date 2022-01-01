from datetime import date
from typing import List, Optional

from pydantic import BaseModel, Field

from pol.api.v0.models.creator import Creator
from pol.api.v0.models.cursorPage import CursorPage
from pol.db.const import TopicStateType


class Post(BaseModel):
    id: int
    content: str  # BBCode
    createdBy: Creator
    createdAt: Optional[date]


class Reply(Post):
    replyTo: Post.id


class Comment(Post):
    replies: Optional[List[Reply]] = Field(max_items=50)
    pagination: Optional[CursorPage]


class Topic(BaseModel):
    id: int
    state: TopicStateType
    title: str
    content: str  # BBCode
    createdBy: Creator
    createdAt: Optional[date]
    comments: List[Comment]
