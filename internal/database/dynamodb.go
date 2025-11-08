package database

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type options struct {
	region   string
	endpoint string
}

type Option func(o *options)

func WithRegion(region string) Option {
	return func(o *options) {
		o.region = region
	}
}

func WithEndpoint(endpoint string) Option {
	return func(o *options) {
		o.endpoint = endpoint
	}
}

func NewDynamoDB(ctx context.Context, opts ...Option) (*dynamodb.Client, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	if o.region == "" {
		return nil, errors.New("region not set")
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(o.region),
	)
	if err != nil {
		return nil, err
	}

	// Set custom endpoint if provided (for local development)
	if o.endpoint != "" {
		cfg.BaseEndpoint = aws.String(o.endpoint)
	}

	client := dynamodb.NewFromConfig(cfg)

	return client, nil
}

// RecoverNil wraps a function that returns an error, catching panics (including nil dereference)
// and returning nil (no error). This allows graceful handling of panics.
func RecoverNil(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Panic recovered, return nil (no error)
			err = nil
		}
	}()
	return fn()
}
