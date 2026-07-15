package service

import (
	"context"
	"testing"
)

func TestLoadDefaultConfigUsesScalingRegionWhenEnvRegionMissing(t *testing.T) {
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_DEFAULT_REGION", "")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")

	cfg, err := loadDefaultConfig(context.Background(), "us-east-1")
	if err != nil {
		t.Fatalf("loadDefaultConfig returned error: %v", err)
	}

	if cfg.Region != "us-east-1" {
		t.Fatalf("expected region us-east-1, got %q", cfg.Region)
	}
}
