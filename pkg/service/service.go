package service

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Service string

const (
	EC2          Service = "ec2"
	Kinesis      Service = "kinesis"
	ElasticCache Service = "elasticache"
	DynamoDB     Service = "dynamodb"

	Unknown Service = "unknown"
)

func getServiceFromString(s string) Service {
	switch s {
	case "ec2":
		return EC2
	case "kinesis":
		return Kinesis
	case "elasticache":
		return ElasticCache
	case "dynamodb":
		return DynamoDB
	default:
		return Unknown
	}
}

func NewConfig(region string, assumeRoleArn string) (*aws.Config, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return nil, err
	}

	creds, err := assumeRoleCreds(ctx, cfg, assumeRoleArn)
	if err != nil {
		return nil, err
	}

	cfg.Credentials = creds
	return &cfg, nil
}

func assumeRoleCreds(ctx context.Context, cfg aws.Config, assumeRoleArn string) (aws.CredentialsProvider, error) {
	stsClient := sts.NewFromConfig(cfg)
	assumeRoleInput := &sts.AssumeRoleInput{
		RoleArn:         aws.String(assumeRoleArn),
		RoleSessionName: aws.String("aws-infra-scaler"),
	}

	result, err := stsClient.AssumeRole(ctx, assumeRoleInput)
	if err != nil {
		return nil, err
	}

	credsProvider := aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     *result.Credentials.AccessKeyId,
			SecretAccessKey: *result.Credentials.SecretAccessKey,
			SessionToken:    *result.Credentials.SessionToken,
			Source:          "AssumeRoleProvider",
		}, nil
	})

	return credsProvider, nil
}

func NewKinesisClient(cfg *aws.Config) kinesis.Client {
	return *kinesis.NewFromConfig(*cfg)
}

func NewElasticCacheClient(cfg *aws.Config) elasticache.Client {
	return *elasticache.NewFromConfig(*cfg)
}

func NewAutoScalingClient(cfg *aws.Config) autoscaling.Client {
	return *autoscaling.NewFromConfig(*cfg)
}
