package service

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/Cool-fire/aws-infra-scaler/pkg/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
)

type mockAutoScalingClient struct {
	updateCalled bool
}

func (m *mockAutoScalingClient) UpdateAutoScalingGroup(ctx context.Context, params *autoscaling.UpdateAutoScalingGroupInput, optFns ...func(*autoscaling.Options)) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	m.updateCalled = true
	return &autoscaling.UpdateAutoScalingGroupOutput{}, nil
}

func TestEC2ScaleServiceDryRunScaleDownDoesNotMutateState(t *testing.T) {
	client := &mockAutoScalingClient{}
	service := EC2Service{Client: client}
	serviceConfig := config.EC2ServiceScalingConfig{
		AsgName:      "test-asg",
		MinCount:     1,
		DesiredCount: 2,
		MaxCount:     3,
	}

	var logBuffer bytes.Buffer
	originalWriter := log.Writer()
	log.SetOutput(&logBuffer)
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
	})

	err := service.ScaleService(context.Background(), serviceConfig, false, true)
	if err != nil {
		t.Fatalf("expected no error for dry-run, got %v", err)
	}

	if client.updateCalled {
		t.Fatal("expected dry-run scale-down to skip UpdateAutoScalingGroup")
	}

	loggedOutput := logBuffer.String()
	if !strings.Contains(loggedOutput, "dry-run: would scale down EC2 Auto Scaling group test-asg") {
		t.Fatalf("expected dry-run log message, got %q", loggedOutput)
	}
}
