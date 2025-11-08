package database

// Type represents the database type
type Type string

const (
	TypeDynamoDB Type = "dynamodb"
	TypePostgres Type = "postgres"
)

// Config holds database-related configuration
type Config struct {
	Type Type
}

