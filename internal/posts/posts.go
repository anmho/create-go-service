package posts

import (
	"time"

	"github.com/google/uuid"
)

func NewPost(
	id uuid.UUID,
	author uuid.UUID,
	title string,
	content string,
	createdAt time.Time) *Post {
	return &Post{
		ID:        id,
		Author:    author,
		Title:     title,
		Content:   content,
		IsPrivate: false,
		CreatedAt: createdAt,
		UpdatedAt: time.Time{},
	}
}

type Post struct {
	ID        uuid.UUID `json:"id" dynamodbav:"id"`
	Author    uuid.UUID `json:"author" dynamodbav:"author"`
	Title     string    `json:"title" dynamodbav:"title"`
	Content   string    `json:"content" dynamodbav:"content"`
	IsPrivate bool      `json:"is_private" dynamodbav:"is_private"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

// ToDynamoDBItem converts a Note to DynamoDB item format

