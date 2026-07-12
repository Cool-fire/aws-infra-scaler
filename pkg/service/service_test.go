package service

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func TestNewConfigUsesAWSProfileWhenSet(t *testing.T) {
	t.Setenv("AWS_PROFILE", "staging-profile")

	originalLoadDefaultConfig := loadDefaultConfig
	originalAssumeRoleCredentials := assumeRoleCredentials
	t.Cleanup(func() {
		loadDefaultConfig = originalLoadDefaultConfig
		assumeRoleCredentials = originalAssumeRoleCredentials
	})

	loadDefaultConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		t.Helper()

		var options config.LoadOptions
		for _, optFn := range optFns {
			if err := optFn(&options); err != nil {
				t.Fatalf("applying load option failed: %v", err)
			}
		}

		if options.SharedConfigProfile != "staging-profile" {
			t.Fatalf("expected shared config profile %q, got %q", "staging-profile", options.SharedConfigProfile)
		}

		return aws.Config{}, nil
	}

	assumeRoleCredentials = func(ctx context.Context, cfg aws.Config, assumeRoleArn string) (aws.CredentialsProvider, error) {
		return aws.AnonymousCredentials{}, nil
	}

	_, err := NewConfig(context.Background(), "us-east-1", "arn:aws:iam::123456789012:role/test")
	if err != nil {
		t.Fatalf("NewConfig returned error: %v", err)
	}
}

func TestNewConfigUsesDefaultCredentialChainWhenAWSProfileUnset(t *testing.T) {
	t.Setenv("AWS_PROFILE", "")

	originalLoadDefaultConfig := loadDefaultConfig
	originalAssumeRoleCredentials := assumeRoleCredentials
	t.Cleanup(func() {
		loadDefaultConfig = originalLoadDefaultConfig
		assumeRoleCredentials = originalAssumeRoleCredentials
	})

	loadDefaultConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		t.Helper()

		var options config.LoadOptions
		for _, optFn := range optFns {
			if err := optFn(&options); err != nil {
				t.Fatalf("applying load option failed: %v", err)
			}
		}

		if options.SharedConfigProfile != "" {
			t.Fatalf("expected no shared config profile, got %q", options.SharedConfigProfile)
		}

		return aws.Config{}, nil
	}

	assumeRoleCredentials = func(ctx context.Context, cfg aws.Config, assumeRoleArn string) (aws.CredentialsProvider, error) {
		return aws.AnonymousCredentials{}, nil
	}

	_, err := NewConfig(context.Background(), "us-east-1", "arn:aws:iam::123456789012:role/test")
	if err != nil {
		t.Fatalf("NewConfig returned error: %v", err)
	}
}
