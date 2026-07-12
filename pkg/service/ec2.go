package service

import (
	"context"
	"fmt"
	"github.com/Cool-fire/aws-infra-scaler/pkg/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
)

type autoScalingAPI interface {
	DescribeAutoScalingGroups(ctx context.Context, params *autoscaling.DescribeAutoScalingGroupsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
	UpdateAutoScalingGroup(ctx context.Context, params *autoscaling.UpdateAutoScalingGroupInput, optFns ...func(*autoscaling.Options)) (*autoscaling.UpdateAutoScalingGroupOutput, error)
}

type EC2Service struct {
	Region string
	Client autoScalingAPI
}

func (ec2 EC2Service) ScaleService(ctx context.Context, ec2ClientConfig config.EC2ServiceScalingConfig) *ScalingError {
	err := validateEc2ScalingConfig(ec2ClientConfig)
	if err != nil {
		return err
	}

	autoScalingGroup, err := ec2.getAutoScalingGroup(ctx, ec2ClientConfig.AsgName)
	if err != nil {
		return &ScalingError{
			ServiceName:  string(EC2),
			IdentifierId: ec2ClientConfig.AsgName,
			Err:          err,
		}
	}

	desiredCapacity := int32(ec2ClientConfig.DesiredCount)
	minSize := int32(ec2ClientConfig.MinCount)
	maxSize := int32(ec2ClientConfig.MaxCount)

	input := autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: autoScalingGroup.AutoScalingGroupName,
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

func (ec2 EC2Service) getAutoScalingGroup(ctx context.Context, asgName string) (*types.AutoScalingGroup, error) {
	autoScalingGroups, err := ec2.listAutoScalingGroups(ctx)
	if err != nil {
		return nil, err
	}

	for _, autoScalingGroup := range autoScalingGroups {
		if autoScalingGroup.AutoScalingGroupName != nil && *autoScalingGroup.AutoScalingGroupName == asgName {
			return &autoScalingGroup, nil
		}
	}

	return nil, fmt.Errorf("auto scaling group %s not found", asgName)
}

func (ec2 EC2Service) listAutoScalingGroups(ctx context.Context) ([]types.AutoScalingGroup, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{}
	autoScalingGroups := make([]types.AutoScalingGroup, 0)

	for {
		output, err := ec2.Client.DescribeAutoScalingGroups(ctx, input)
		if err != nil {
			return nil, err
		}

		autoScalingGroups = append(autoScalingGroups, output.AutoScalingGroups...)
		if output.NextToken == nil || *output.NextToken == "" {
			return autoScalingGroups, nil
		}

		input.NextToken = output.NextToken
	}
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
