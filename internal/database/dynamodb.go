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

func sum(name string, nums ...int) {

}

func NewDynamoDB(ctx context.Context, opts ...Option) (*dynamodb.Client, error) {
	sum("s", []int{1, 2, 3}...)
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	if o.region == "" {
		return nil, errors.New("region not set")
	}

	var cfg aws.Config
	var err error
	if o.endpoint != "" {
		cfg, err = awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(o.region),
			awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL:           o.endpoint,
						SigningRegion: o.region,
					}, nil
				},
			)))
	} else {
		cfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(o.region))
	}

	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(cfg)

	return client, nil
}
