package service

import (
	"context"
	"fmt"
	"github.com/Cool-fire/aws-infra-scaler/pkg/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
)

type autoScalingClient interface {
	UpdateAutoScalingGroup(ctx context.Context, params *autoscaling.UpdateAutoScalingGroupInput, optFns ...func(*autoscaling.Options)) (*autoscaling.UpdateAutoScalingGroupOutput, error)
	DescribeAutoScalingGroups(ctx context.Context, params *autoscaling.DescribeAutoScalingGroupsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
}

type EC2Service struct {
	Region string
	Client autoScalingClient
}

func (ec2 EC2Service) ScaleService(ctx context.Context, ec2ClientConfig config.EC2ServiceScalingConfig) *ScalingError {
	err := validateEc2ScalingConfig(ec2ClientConfig)
	if err != nil {
		return err
	}

	desiredCapacity := int32(ec2ClientConfig.DesiredCount)
	minSize := int32(ec2ClientConfig.MinCount)
	maxSize := int32(ec2ClientConfig.MaxCount)

	input := autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: &ec2ClientConfig.AsgName,
		DesiredCapacity:      &desiredCapacity,
		MaxSize:              &maxSize,
		MinSize:              &minSize,
	}

	_, scaleError := ec2.Client.UpdateAutoScalingGroup(ctx, &input)

	if scaleError != nil {
		return &ScalingError{
			ServiceName:  string(EC2),
			IdentifierId: ec2ClientConfig.AsgName,
			Err:          scaleError,
		}
	}

	return nil
}

func validateEc2ScalingConfig(clientConfig config.EC2ServiceScalingConfig) *ScalingError {
	if clientConfig.AsgName == "" || clientConfig.DesiredCount <= 0 || clientConfig.MaxCount <= 0 || clientConfig.MinCount <= 0 {
		return &ScalingError{
			ServiceName:  string(EC2),
			IdentifierId: clientConfig.AsgName,
			Err:          fmt.Errorf("invalid scaling config"),
		}
	}
	return nil
}

func (ec2 EC2Service) getAutoScalingGroup(ctx context.Context, asgName string) (*types.AutoScalingGroup, *ScalingError) {
	output, err := ec2.Client.DescribeAutoScalingGroups(ctx, &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{asgName},
	})
	if err != nil {
		return nil, &ScalingError{
			ServiceName:  string(EC2),
			IdentifierId: asgName,
			Err:          err,
		}
	}

	if len(output.AutoScalingGroups) == 0 {
		return nil, &ScalingError{
			ServiceName:  string(EC2),
			IdentifierId: asgName,
			Err:          fmt.Errorf("auto scaling group not found"),
		}
	}

	return &output.AutoScalingGroups[0], nil
}

func countAutoScalingGroupCapacity(group *types.AutoScalingGroup) int32 {
	if group == nil {
		return 0
	}

	seen := make(map[string]struct{}, len(group.Instances))
	var capacity int32

	for _, instance := range group.Instances {
		if instance.InstanceId == nil || !isCapacityContributingLifecycleState(string(instance.LifecycleState)) {
			continue
		}

		instanceID := *instance.InstanceId
		if _, ok := seen[instanceID]; ok {
			continue
		}

		seen[instanceID] = struct{}{}
		capacity++
	}

	return capacity
}

func isCapacityContributingLifecycleState(state string) bool {
	switch state {
	case "Pending", "Pending:Wait", "Pending:Proceed", "InService":
		return true
	default:
		return false
	}
}
