package service

import (
	"context"
	"github.com/Cool-fire/aws-infra-scaler/pkg/config"
	"github.com/Cool-fire/aws-infra-scaler/pkg/errors"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
)

func ScaleEc2Service(ctx context.Context, ec2ClientConfig config.EC2ServiceScalingConfig, client autoscaling.Client) *errors.ScalingFailureError {
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

	_, scaleError := client.UpdateAutoScalingGroup(ctx, &input)

	if scaleError != nil {
		return &errors.ScalingFailureError{
			ServiceName:  EC2,
			IdentifierId: ec2ClientConfig.AsgName,
			Reason:       scaleError.Error(),
		}
	}

	return nil
}

func validateEc2ScalingConfig(clientConfig config.EC2ServiceScalingConfig) *errors.ScalingFailureError {
	if clientConfig.AsgName == "" || clientConfig.DesiredCount <= 0 || clientConfig.MaxCount <= 0 || clientConfig.MinCount <= 0 {
		return &errors.ScalingFailureError{
			ServiceName:  EC2,
			IdentifierId: clientConfig.AsgName,
			Reason:       "Invalid scaling config",
		}
	}
	return nil
}
