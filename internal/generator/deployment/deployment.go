package deployment

// Type represents the deployment type
type Type string

const (
	TypeFly Type = "fly"
)

// Config holds deployment-related configuration
type Config struct {
	Type Type
}

