package service

import (
	"context"
	"fmt"
	"github.com/Cool-fire/aws-infra-scaler/pkg/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
)

type EC2Service struct {
	Region string
	Client *autoscaling.Client
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
