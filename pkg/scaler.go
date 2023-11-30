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

type ScalingResponse struct {
	ContainsFailedServices bool
	RegionalFailedServices map[string][]*service.ScalingError
}

var assumeRoleArn string

func ScaleApp(shouldScaleUp bool, configPath string) (*ScalingResponse, error) {

	scalingConfig, err := config.ReadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	assumeRoleArn = scalingConfig.AssumedRoleArn
	if assumeRoleArn == "" {
		return nil, errors.New("no assumed role ARN provided")
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

	regionalFailedServices := make(map[string][]*service.ScalingError)
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled")
		case result, open := <-resultChan:

			if !open {
				if len(regionalFailedServices) > 0 {
					return &ScalingResponse{
						ContainsFailedServices: true,
						RegionalFailedServices: regionalFailedServices,
					}, nil
				} else {
					return &ScalingResponse{
						ContainsFailedServices: false,
						RegionalFailedServices: nil,
					}, nil
				}
			}

			if result != nil {
				if regionalFailedServices[result.Region] == nil {
					regionalFailedServices[result.Region] = make([]*service.ScalingError, 0)
				}
				regionalFailedServices[result.Region] = append(regionalFailedServices[result.Region], result)
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
			err.Region = region
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
			err.Region = region
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
			err.Region = region
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
				err.Region = region
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
