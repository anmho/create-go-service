package notes

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
)

type Service struct {
	ddbClient *dynamodb.Client
	tableName string
}

func NewService(ddbClient *dynamodb.Client, tableName string) *Service {
	return &Service{
		ddbClient: ddbClient,
		tableName: tableName,
	}
}

func (s *Service) CreateNote(
	ctx context.Context,
	id uuid.UUID,
	author uuid.UUID,
	content string,
	title string) (*Note, error) {

	note := NewNote(id, author, title, content, time.Now())
	item, err := NoteToDynamoItem(note)
	if err != nil {
		return nil, err
	}

	result, err := s.ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})

	if err != nil {
		return nil, err
	}

	fmt.Println(result)

	return note, nil
}

func (s *Service) GetNote(id uuid.UUID) {

}

func (s *Service) ListNotes() {

}

func (s *Service) DeleteNote(id uuid.UUID) {

}
