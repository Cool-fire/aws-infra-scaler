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
		go scaleService(ctx, awsCreds, serviceScaleConfig, shouldScaleUp, scalingRegion.Region, &serviceWg, resultChan)
	}

	serviceWg.Wait()
	fmt.Println("done scaling region ", scalingRegion.Region)
}

func scaleService(ctx context.Context, awsCreds *aws.Config, serviceScaleConfig interface{}, shouldScaleUp bool, region string, wg *sync.WaitGroup, resultChan chan error) {
	defer wg.Done()

	switch serviceScaleConfig.(type) {
	case config.KinesisServiceScalingConfig:
		kinesisClient := service.NewKinesisClient(awsCreds)
		kinesisClientConfig := serviceScaleConfig.(config.KinesisServiceScalingConfig)

		ks := service.KinesisService{
			Region: region,
			Client: kinesisClient,
		}
		err := ks.ScaleService(ctx, kinesisClientConfig)
		if err != nil {
			fmt.Println("Error scaling Kinesis service: ", err)
			resultChan <- err
		}

	case config.EC2ServiceScalingConfig:
		autoScalingClient := service.NewAutoScalingClient(awsCreds)
		ec2ClientConfig := serviceScaleConfig.(config.EC2ServiceScalingConfig)

		ec2 := service.EC2Service{
			Region: region,
			Client: autoScalingClient,
		}

		err := ec2.ScaleService(ctx, ec2ClientConfig)
		if err != nil {
			fmt.Println("Error scaling EC2 service: ", err)
			resultChan <- err
		}

	case config.ElasticCacheServiceScalingConfig:
		fmt.Println("Scaling ElasticCache")
		elasticCacheClient := service.NewElasticCacheClient(awsCreds)
		elasticCacheClientConfig := serviceScaleConfig.(config.ElasticCacheServiceScalingConfig)

		es := service.ElasticCacheService{
			Region: region,
			Client: elasticCacheClient,
		}

		err := es.ScaleService(ctx, elasticCacheClientConfig, shouldScaleUp)
		if err != nil {
			fmt.Println("Error scaling ElasticCache service: ", err)
			resultChan <- err
		}

	case config.DynamoDBServiceScalingConfig:
		fmt.Println("Scaling DynamoDB")
		appAutoScalingClient := service.NewApplicationAutoScalingClient(awsCreds)
		dynamoDBClientConfig := serviceScaleConfig.(config.DynamoDBServiceScalingConfig)

		ds := service.DynamoDBService{
			Region: region,
			Client: appAutoScalingClient,
		}

		errs := ds.ScaleService(ctx, dynamoDBClientConfig)
		if errs != nil {
			fmt.Println("Error scaling DynamoDB service: ", errs)
			for _, err := range errs {
				resultChan <- err
			}
		}

	default:
		fmt.Println("Unknown service")
		resultChan <- &errors.ScalingFailureError{
			ServiceName:  "Unknown",
			IdentifierId: "Unknown",
			Reason:       "Unknown service",
		}
	}
}
