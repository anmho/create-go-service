package notes

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

func NewNote(
	id uuid.UUID,
	author uuid.UUID,
	title string,
	content string,
	createdAt time.Time) *Note {
	return &Note{
		ID:        id,
		Author:    author,
		Title:     title,
		Content:   content,
		IsPrivate: false,
		CreatedAt: createdAt,
		UpdatedAt: time.Time{},
	}
}

type Note struct {
	ID        uuid.UUID `json:"id" dynamodbav:"id"`
	Author    uuid.UUID `json:"author" dynamodbav:"author"`
	Title     string    `json:"title" dynamodbav:"title"`
	Content   string    `json:"content" dynamodbav:"content"`
	IsPrivate bool      `json:"is_private" dynamodbav:"is_private"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

// ToDynamoDBItem converts a Note to DynamoDB item format
func NoteToDynamoItem(note *Note) (map[string]types.AttributeValue, error) {
	return attributevalue.MarshalMap(note)
}

// NoteFromDynamoItem converts DynamoDB item to Note
func NoteFromDynamoItem(item map[string]types.AttributeValue) (*Note, error) {
	var note Note
	err := attributevalue.UnmarshalMap(item, &note)
	return &note, err
}
