package dynamo

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Config struct {
	Region    string
	Endpoint  string
	TableName string
}

func NewClient(ctx context.Context, cfg Config) (*dynamodb.Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	opts := []func(*dynamodb.Options){}
	if cfg.Endpoint != "" {
		opts = append(opts, func(o *dynamodb.Options) {
			o.BaseEndpoint = &cfg.Endpoint
		})
	}

	return dynamodb.NewFromConfig(awsCfg, opts...), nil
}
