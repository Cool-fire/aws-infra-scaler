package service

import (
	"context"
	"github.com/Cool-fire/aws-infra-scaler/pkg/config"
	"github.com/Cool-fire/aws-infra-scaler/pkg/errors"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
)

func ScaleKinesisService(ctx context.Context, kinesisClientConfig config.KinesisServiceScalingConfig, client kinesis.Client) *errors.ScalingFailureError {
	err := validateKinesisScalingConfig(kinesisClientConfig)
	if err != nil {
		return err
	}

	targetShareCount := int32(kinesisClientConfig.DesiredShardCount)
	input := kinesis.UpdateShardCountInput{
		StreamName:       &kinesisClientConfig.StreamArn,
		TargetShardCount: &targetShareCount,
		ScalingType:      types.ScalingTypeUniformScaling,
	}

	_, scaleError := client.UpdateShardCount(ctx, &input)
	if scaleError != nil {
		return &errors.ScalingFailureError{
			ServiceName:  Kinesis,
			IdentifierId: kinesisClientConfig.StreamArn,
			Reason:       scaleError.Error(),
		}
	}
	return nil
}

func validateKinesisScalingConfig(clientConfig config.KinesisServiceScalingConfig) *errors.ScalingFailureError {
	if clientConfig.StreamArn == "" || clientConfig.DesiredShardCount <= 0 {
		return &errors.ScalingFailureError{
			ServiceName:  Kinesis,
			IdentifierId: clientConfig.StreamArn,
			Reason:       "Invalid scaling config",
		}
	}
	return nil
}
