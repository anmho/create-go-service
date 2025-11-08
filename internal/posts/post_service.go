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


type PostParams struct {
	AuthorID uuid.UUID
	Title string
	Content string
}

func (s *Service) CreatePost(
	ctx context.Context,
	postId uuid.UUID,
	params PostParams,
	) (*Post, error) {

	post := NewPost(
		postId, 
		params.AuthorID,
		params.Title, 
		params.Content, 
		time.Now())
	err := s.postTable.Put(ctx, post)

	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *Service) UpdatePost(
	ctx context.Context,
	postId uuid.UUID,
	params PostParams,
) (*Post, error) {
	// First get the existing post to preserve metadata
	existingPost, err := s.GetPost(ctx, postId)
	if err != nil {
		return nil, err
	}

	// Update the post with new data while preserving metadata
	updatedPost := &Post{
		ID:        existingPost.ID,
		Author:    existingPost.Author,
		Title:     params.Title,
		Content:   params.Content,
		IsPrivate: false, // Default to false, could be made configurable
		CreatedAt: existingPost.CreatedAt,
		UpdatedAt: time.Now(),
	}

	err = s.postTable.Put(ctx, updatedPost)
	if err != nil {
		return nil, err
	}

	return updatedPost, nil
}

func (s *Service) GetPost(ctx context.Context, postId uuid.UUID) (*Post, error) {
	return s.postTable.Get(ctx, postId)
}

func (s *Service) ListPosts(ctx context.Context, authorId uuid.UUID) ([]Post, error) {
	return s.postTable.List(ctx, authorId)
}

func (s *Service) DeletePost(ctx context.Context, id uuid.UUID) error {
	return s.postTable.Delete(ctx, id)
}
