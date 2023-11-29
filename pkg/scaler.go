package pkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/Cool-fire/aws-infra-scaler/pkg/config"
	"github.com/Cool-fire/aws-infra-scaler/pkg/service"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/smithy-go"
	"sync"
)

type ScalingResult struct {
	region  string
	service string
	err     error
}

var assumeRoleArn string

func ScaleApp(shouldScaleUp bool, configPath string) error {
	scalingConfig, err := config.ReadConfig(configPath)
	assumeRoleArn = scalingConfig.AssumedRoleArn

	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan *service.ScalingError)

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
				return nil
			}

			var se *service.ScalingError
			if errors.As(result, &se) {
				fmt.Printf("Scaling error: %+v\n", se)
			} else {
				panic(fmt.Sprintf("unknown error: %+v", result))
			}
		}
	}
}

func scaleRegion(ctx context.Context, scalingRegion config.ScalingRegion, shouldScaleUp bool, wg *sync.WaitGroup, resultChan chan *service.ScalingError) {
	defer wg.Done()

	var serviceWg sync.WaitGroup
	awsCreds, err := service.NewConfig(ctx, scalingRegion.Region, assumeRoleArn)
	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			resultChan <- &service.ScalingError{
				ServiceName:  oe.Service(),
				IdentifierId: oe.ServiceID,
				Err:          oe,
			}
		}
		return
	}

	for _, serviceScaleConfig := range scalingRegion.ServiceScaleConfigs {
		serviceWg.Add(1)
		go scaleService(ctx, awsCreds, serviceScaleConfig, shouldScaleUp, scalingRegion.Region, &serviceWg, resultChan)
	}

	serviceWg.Wait()
}

func scaleService(ctx context.Context, awsCreds *aws.Config, serviceScaleConfig interface{}, shouldScaleUp bool, region string, wg *sync.WaitGroup, resultChan chan *service.ScalingError) {
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
			resultChan <- err
		}

	case config.ElasticCacheServiceScalingConfig:
		elasticCacheClient := service.NewElasticCacheClient(awsCreds)
		elasticCacheClientConfig := serviceScaleConfig.(config.ElasticCacheServiceScalingConfig)

		es := service.ElasticCacheService{
			Region: region,
			Client: elasticCacheClient,
		}

		err := es.ScaleService(ctx, elasticCacheClientConfig, shouldScaleUp)
		if err != nil {
			resultChan <- err
		}

	case config.DynamoDBServiceScalingConfig:
		appAutoScalingClient := service.NewApplicationAutoScalingClient(awsCreds)
		dynamoDBClientConfig := serviceScaleConfig.(config.DynamoDBServiceScalingConfig)

		ds := service.DynamoDBService{
			Region: region,
			Client: appAutoScalingClient,
		}

		errs := ds.ScaleService(ctx, dynamoDBClientConfig)
		if errs != nil {
			for _, err := range errs {
				resultChan <- err
			}
		}

	default:
		resultChan <- &service.ScalingError{
			ServiceName:  "Unknown",
			IdentifierId: "Unknown",
			Err:          fmt.Errorf("unknown service"),
		}
	}
}
