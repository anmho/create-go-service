package posts

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	postTable *Table
}

func NewService(postTable *Table) *Service {
	return &Service{
		postTable: postTable,
	}
}

func (s *Service) CreateNote(
	ctx context.Context,
	id uuid.UUID,
	author uuid.UUID,
	content string,
	title string) (*Post, error) {

	note := NewPost(id, author, title, content, time.Now())

	err := s.postTable.Put(ctx, note)

	if err != nil {
		return nil, err
	}

	return note, nil
}

func (s *Service) GetNote(ctx context.Context, postId uuid.UUID) (*Post, error){
	return s.postTable.Get(ctx, postId)
}

func (s *Service) ListNotes(ctx context.Context, creatorID string) ([]Post, error){
	return s.postTable.List(ctx, creatorID)
}

func (s *Service) DeleteNote(ctx context.Context, id uuid.UUID) error {
	return s.postTable.Delete(ctx, id)
}
