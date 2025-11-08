package posts

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const PostTableName string = "PostTable"
const UserID string = "user_id"
const CreatedAt string = "created_at"
const PostID string = "post_id"
const PostIDGSI string = "GSI_PostID"
const DefaultRCU = 100
const DefaultWCU = 100

var PostNotFoundError error = errors.New("Post not found")

// Table is a repository for DynamoDB operations on posts
type Table struct {
	dynamoClient *dynamodb.Client
}

// NewTable creates a new posts table repository and tests the connection
// by describing the table to fail fast if the connection fails
func NewTable(ctx context.Context, dynamoClient *dynamodb.Client) (*Table, error) {
	// Test connection by describing the table - fail fast if connection fails
	_, err := dynamoClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(PostTableName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DynamoDB table %s: %w", PostTableName, err)
	}

	return &Table{
		dynamoClient: dynamoClient,
	}, nil
}

func (t *Table) Put(ctx context.Context, post *Post) error {
	valueMap, err := attributevalue.MarshalMap(post)
	if err != nil {
		return fmt.Errorf("error during PUT to %s %w", PostTableName, err)
	}
	_, err = t.dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		Item:                   valueMap,
		TableName:              aws.String(PostTableName),
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	})
	if err != nil {
		return err
	}
	return nil
}

// List returns all posts authored by the user with id userID
func (t *Table) List(ctx context.Context, userID uuid.UUID) ([]Post, error) {
	params := &dynamodb.QueryInput{
		TableName:              aws.String(PostTableName),
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :userID", UserID)),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{Value: userID.String()},
		},
		ConsistentRead:         aws.Bool(false),
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
		ScanIndexForward:       aws.Bool(false),
	}

	result, err := t.dynamoClient.Query(ctx, params)
	if err != nil {
		return nil, err
	}
	posts := []Post{}

	err = attributevalue.UnmarshalListOfMaps(result.Items, &posts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal posts: %w", err)
	}

	return posts, nil
}

// Get retrieves a post by its ID using the GSI_PostID index
func (t *Table) Get(ctx context.Context, postID uuid.UUID) (*Post, error) {
	params := &dynamodb.QueryInput{
		TableName:              aws.String(PostTableName),
		IndexName:              aws.String(PostIDGSI),
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :postID", PostID)),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":postID": &types.AttributeValueMemberS{Value: postID.String()},
		},
		ConsistentRead:         aws.Bool(false),
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	}

	result, err := t.dynamoClient.Query(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query post by ID %s: %w", postID, err)
	}

	if len(result.Items) == 0 {
		return nil, PostNotFoundError
	}

	if len(result.Items) > 1 {
		return nil, fmt.Errorf("multiple posts found with ID %s", postID)
	}

	var post Post
	err = attributevalue.UnmarshalMap(result.Items[0], &post)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal post: %w", err)
	}

	return &post, nil
}

// DeleteByPostId removes a post by post ID (finds it via GSI first, then deletes from main table)
func (t *Table) Delete(ctx context.Context, postID uuid.UUID) error {
	// First query GSI to get the post and its primary key components
	post, err := t.Get(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to find post with ID %s for deletion: %w", postID.String(), err)
	}

	// Delete from main table using primary key components
	params := &dynamodb.DeleteItemInput{
		TableName: aws.String(PostTableName),
		Key: map[string]types.AttributeValue{
			UserID:    &types.AttributeValueMemberS{Value: post.Author.String()},
			CreatedAt: &types.AttributeValueMemberS{Value: post.CreatedAt.Format(time.RFC3339)},
		},
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	}

	_, err = t.dynamoClient.DeleteItem(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}
