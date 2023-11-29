package service

import (
	"context"
	"fmt"
	"github.com/Cool-fire/aws-infra-scaler/pkg/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
)

type KinesisService struct {
	Region string
	Client *kinesis.Client
}

func (k KinesisService) ScaleService(ctx context.Context, kinesisServiceScalingConfig config.KinesisServiceScalingConfig) *ScalingError {
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
		return &ScalingError{
			ServiceName:  string(Kinesis),
			IdentifierId: kinesisServiceScalingConfig.StreamArn,
			Err:          scaleError,
		}
	}
	return nil
}

func validateKinesisScalingConfig(clientConfig config.KinesisServiceScalingConfig) *ScalingError {
	if clientConfig.StreamArn == "" || clientConfig.DesiredShardCount <= 0 {
		return &ScalingError{
			ServiceName:  string(Kinesis),
			IdentifierId: clientConfig.StreamArn,
			Err:          fmt.Errorf("invalid scaling config"),
		}
	}
	return nil
}
