package service

import (
	"context"
	"testing"

	"github.com/Cool-fire/aws-infra-scaler/pkg/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
)

type fakeAutoScalingClient struct {
	describeOutputs []*autoscaling.DescribeAutoScalingGroupsOutput
	describeInputs  []*autoscaling.DescribeAutoScalingGroupsInput
	updateInput     *autoscaling.UpdateAutoScalingGroupInput
}

func (f *fakeAutoScalingClient) DescribeAutoScalingGroups(_ context.Context, params *autoscaling.DescribeAutoScalingGroupsInput, _ ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	cloned := *params
	f.describeInputs = append(f.describeInputs, &cloned)

	output := f.describeOutputs[0]
	f.describeOutputs = f.describeOutputs[1:]
	return output, nil
}

func (f *fakeAutoScalingClient) UpdateAutoScalingGroup(_ context.Context, params *autoscaling.UpdateAutoScalingGroupInput, _ ...func(*autoscaling.Options)) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	f.updateInput = params
	return &autoscaling.UpdateAutoScalingGroupOutput{}, nil
}

func TestEC2ServiceListAutoScalingGroupsFetchesAllPages(t *testing.T) {
	client := &fakeAutoScalingClient{
		describeOutputs: []*autoscaling.DescribeAutoScalingGroupsOutput{
			{
				AutoScalingGroups: []types.AutoScalingGroup{{AutoScalingGroupName: aws.String("asg-1")}},
				NextToken:         aws.String("page-2"),
			},
			{
				AutoScalingGroups: []types.AutoScalingGroup{{AutoScalingGroupName: aws.String("asg-2")}},
			},
		},
	}

	ec2 := EC2Service{Client: client}
	autoScalingGroups, err := ec2.listAutoScalingGroups(context.Background())
	if err != nil {
		t.Fatalf("listAutoScalingGroups returned error: %v", err)
	}

	if len(autoScalingGroups) != 2 {
		t.Fatalf("expected 2 auto scaling groups, got %d", len(autoScalingGroups))
	}

	if len(client.describeInputs) != 2 {
		t.Fatalf("expected 2 describe calls, got %d", len(client.describeInputs))
	}

	if client.describeInputs[0].NextToken != nil {
		t.Fatalf("expected first describe call without next token, got %v", *client.describeInputs[0].NextToken)
	}

	if client.describeInputs[1].NextToken == nil || *client.describeInputs[1].NextToken != "page-2" {
		t.Fatalf("expected second describe call to use next token page-2, got %v", client.describeInputs[1].NextToken)
	}

	if autoScalingGroups[1].AutoScalingGroupName == nil || *autoScalingGroups[1].AutoScalingGroupName != "asg-2" {
		t.Fatalf("expected second auto scaling group to be asg-2, got %v", autoScalingGroups[1].AutoScalingGroupName)
	}
}

func TestEC2ServiceScaleServiceUsesAutoScalingGroupFromLaterPage(t *testing.T) {
	client := &fakeAutoScalingClient{
		describeOutputs: []*autoscaling.DescribeAutoScalingGroupsOutput{
			{
				AutoScalingGroups: []types.AutoScalingGroup{{AutoScalingGroupName: aws.String("asg-1")}},
				NextToken:         aws.String("page-2"),
			},
			{
				AutoScalingGroups: []types.AutoScalingGroup{{AutoScalingGroupName: aws.String("asg-200")}},
			},
		},
	}

	ec2 := EC2Service{Client: client}
	err := ec2.ScaleService(context.Background(), config.EC2ServiceScalingConfig{
		AsgName:      "asg-200",
		DesiredCount: 2,
		MinCount:     1,
		MaxCount:     3,
	})
	if err != nil {
		t.Fatalf("ScaleService returned error: %v", err)
	}

	if client.updateInput == nil {
		t.Fatal("expected UpdateAutoScalingGroup to be called")
	}

	if client.updateInput.AutoScalingGroupName == nil || *client.updateInput.AutoScalingGroupName != "asg-200" {
		t.Fatalf("expected UpdateAutoScalingGroup to target asg-200, got %v", client.updateInput.AutoScalingGroupName)
	}

	if client.updateInput.DesiredCapacity == nil || *client.updateInput.DesiredCapacity != 2 {
		t.Fatalf("expected desired capacity 2, got %v", client.updateInput.DesiredCapacity)
	}
	if client.updateInput.MinSize == nil || *client.updateInput.MinSize != 1 {
		t.Fatalf("expected min size 1, got %v", client.updateInput.MinSize)
	}
	if client.updateInput.MaxSize == nil || *client.updateInput.MaxSize != 3 {
		t.Fatalf("expected max size 3, got %v", client.updateInput.MaxSize)
	}
}
