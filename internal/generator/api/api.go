package api

// Type represents the API framework type
type Type string

const (
	TypeChi  Type = "chi"
	TypeHuma Type = "huma"
	TypeGRPC Type = "grpc"
)

// Config holds API-related configuration
type Config struct {
	Types []Type // API types to generate (chi, grpc, huma)
}
