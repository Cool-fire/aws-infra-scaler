package service

import (
	"context"
	"github.com/Cool-fire/aws-infra-scaler/pkg/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
)

type KinesisService struct {
	Region string
	Client kinesis.Client
}

func (k KinesisService) ScaleService(ctx context.Context, kinesisServiceScalingConfig config.KinesisServiceScalingConfig) *ScalingFailureError {
	err := validateKinesisScalingConfig(kinesisServiceScalingConfig)
	if err != nil {
		return err
	}

	targetShareCount := int32(kinesisServiceScalingConfig.DesiredShardCount)
	input := kinesis.UpdateShardCountInput{
		StreamName:       &kinesisServiceScalingConfig.StreamArn,
		TargetShardCount: &targetShareCount,
		ScalingType:      types.ScalingTypeUniformScaling,
	}

	_, scaleError := k.Client.UpdateShardCount(ctx, &input)
	if scaleError != nil {
		return &ScalingFailureError{
			ServiceName:  Kinesis,
			IdentifierId: kinesisServiceScalingConfig.StreamArn,
			Reason:       scaleError.Error(),
		}
	}
	return nil
}

func validateKinesisScalingConfig(clientConfig config.KinesisServiceScalingConfig) *ScalingFailureError {
	if clientConfig.StreamArn == "" || clientConfig.DesiredShardCount <= 0 {
		return &ScalingFailureError{
			ServiceName:  Kinesis,
			IdentifierId: clientConfig.StreamArn,
			Reason:       "Invalid scaling config",
		}
	}
	return nil
}
