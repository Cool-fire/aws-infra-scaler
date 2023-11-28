package pkg

import (
	"context"
	"fmt"
	"github.com/Cool-fire/aws-infra-scaler/pkg/config"
	"github.com/Cool-fire/aws-infra-scaler/pkg/errors"
	"github.com/Cool-fire/aws-infra-scaler/pkg/service"
	"github.com/aws/aws-sdk-go-v2/aws"
	"sync"
)

type ScalingResult struct {
	region  string
	service string
	err     error
}

var assumeRoleArn string

func ScaleApplication(shouldScaleUp bool, configPath string) error {
	scalingConfig, err := config.ReadConfig(configPath)
	assumeRoleArn = scalingConfig.AssumedRoleArn

	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan error)

	go func() {
		defer close(resultChan)
		var wg sync.WaitGroup
		for _, scalingRegion := range scalingConfig.ScalingRegions {
			wg.Add(1)
			go scaleRegion(ctx, scalingRegion, shouldScaleUp, &wg, resultChan)
		}
		wg.Wait()
	}()

	fmt.Println("Waiting for results...")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context cancelled")
			return nil
		case result, ok := <-resultChan:
			if !ok {
				fmt.Println("Result channel closed")
				return nil
			}
			fmt.Println("Result received:", result)
		}
	}
}

func scaleRegion(ctx context.Context, scalingRegion config.ScalingRegion, shouldScaleUp bool, wg *sync.WaitGroup, resultChan chan error) {
	defer wg.Done()

	var serviceWg sync.WaitGroup
	awsCreds, _ := service.NewConfig(scalingRegion.Region, assumeRoleArn)

	for _, serviceScaleConfig := range scalingRegion.ServiceScaleConfigs {
		serviceWg.Add(1)
		go scaleService(ctx, awsCreds, serviceScaleConfig, shouldScaleUp, &serviceWg, resultChan)
	}

	serviceWg.Wait()
	fmt.Println("done scaling region ", scalingRegion.Region)
}

func scaleService(ctx context.Context, awsCreds *aws.Config, serviceScaleConfig interface{}, shouldScaleUp bool, wg *sync.WaitGroup, resultChan chan error) {
	defer wg.Done()

	switch serviceScaleConfig.(type) {
	case config.KinesisServiceScalingConfig:
		kinesisClient := service.NewKinesisClient(awsCreds)
		kinesisClientConfig := serviceScaleConfig.(config.KinesisServiceScalingConfig)
		err := service.ScaleKinesisService(ctx, kinesisClientConfig, kinesisClient)
		if err != nil {
			fmt.Println("Error scaling Kinesis service: ", err)
			resultChan <- err
		}

	case config.EC2ServiceScalingConfig:
		autoScalingClient := service.NewAutoScalingClient(awsCreds)
		ec2ClientConfig := serviceScaleConfig.(config.EC2ServiceScalingConfig)
		err := service.ScaleEc2Service(ctx, ec2ClientConfig, autoScalingClient)
		if err != nil {
			fmt.Println("Error scaling EC2 service: ", err)
			resultChan <- err
		}

	case config.ElasticCacheServiceScalingConfig:
		fmt.Println("Scaling ElasticCache")
		resultChan <- nil

	case config.DynamoDBServiceScalingConfig:
		fmt.Println("Scaling DynamoDB")
		resultChan <- nil

	default:
		fmt.Println("Unknown service")
		resultChan <- &errors.ScalingFailureError{
			ServiceName:  "Unknown",
			IdentifierId: "Unknown",
			Reason:       "Unknown service",
		}
	}
}
