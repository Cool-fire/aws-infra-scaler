package service

import (
	"context"
	"fmt"
	"github.com/Cool-fire/aws-infra-scaler/pkg/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"log"
)

type autoScalingAPI interface {
	UpdateAutoScalingGroup(ctx context.Context, params *autoscaling.UpdateAutoScalingGroupInput, optFns ...func(*autoscaling.Options)) (*autoscaling.UpdateAutoScalingGroupOutput, error)
}

type EC2Service struct {
	Region string
	Client autoScalingAPI
}

func (ec2 EC2Service) ScaleService(ctx context.Context, ec2ClientConfig config.EC2ServiceScalingConfig, shouldScaleUp bool, dryRun bool) *ScalingError {
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

	action := "scale up"
	if !shouldScaleUp {
		action = "scale down"
	}

	if dryRun {
		log.Printf("dry-run: would %s EC2 Auto Scaling group %s to min=%d desired=%d max=%d", action, ec2ClientConfig.AsgName, ec2ClientConfig.MinCount, ec2ClientConfig.DesiredCount, ec2ClientConfig.MaxCount)
		return nil
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
