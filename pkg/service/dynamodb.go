package service

import (
	"context"
	"fmt"
	"github.com/Cool-fire/aws-infra-scaler/pkg/config"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
	"sync"
)

const (
	DynamodbServiceNamespace = "dynamodb"
)

var applicationAutoscalingClient *applicationautoscaling.Client

type DynamoDBService struct {
	Region string
	Client *applicationautoscaling.Client
}

func (ds DynamoDBService) ScaleService(ctx context.Context, dynamodbClientConfig config.DynamoDBServiceScalingConfig) []*ScalingError {
	err := validateDynamoDBScalingConfig(dynamodbClientConfig)
	if err != nil {
		return []*ScalingError{err}
	}
	applicationAutoscalingClient = ds.Client

	errChan := make(chan *ScalingError)

	go scaleDynamoDB(&ctx, dynamodbClientConfig, errChan)

	var scalingErrors []*ScalingError
	for err := range errChan {
		if err != nil {
			scalingErrors = append(scalingErrors, err)
		}
	}

	if len(scalingErrors) > 0 {
		return scalingErrors
	} else {
		return nil
	}
}

func scaleDynamoDB(ctx *context.Context, dynamodbClientConfig config.DynamoDBServiceScalingConfig, errChan chan<- *ScalingError) {
	defer close(errChan)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := scaleRCU(*ctx, dynamodbClientConfig.RCU, dynamodbClientConfig.IsIndex, dynamodbClientConfig.TableName, &wg)
		errChan <- err
	}()

	wg.Add(1)
	go func() {
		err := scaleWCU(*ctx, dynamodbClientConfig.WCU, dynamodbClientConfig.IsIndex, dynamodbClientConfig.TableName, &wg)
		errChan <- err
	}()

	wg.Wait()
}

func scaleRCU(ctx context.Context, rcu config.RCU, isIndex bool, tableName string, wg *sync.WaitGroup) *ScalingError {
	defer wg.Done()

	scalableDimension := types.ScalableDimensionDynamoDBTableReadCapacityUnits
	if isIndex {
		scalableDimension = types.ScalableDimensionDynamoDBIndexReadCapacityUnits
	}
	err := scaleDB(ctx, scalableDimension, tableName, int32(rcu.MinProvisionedCapacity), int32(rcu.MaxProvisionedCapacity))
	if err != nil {
		return err
	}

	return nil
}

func scaleWCU(ctx context.Context, wcu config.WCU, isIndex bool, tableName string, wg *sync.WaitGroup) *ScalingError {
	defer wg.Done()

	scalableDimension := types.ScalableDimensionDynamoDBTableWriteCapacityUnits
	if isIndex {
		scalableDimension = types.ScalableDimensionDynamoDBIndexWriteCapacityUnits
	}
	err := scaleDB(ctx, scalableDimension, tableName, int32(wcu.MinProvisionedCapacity), int32(wcu.MaxProvisionedCapacity))
	if err != nil {
		return err
	}

	return nil
}

func scaleDB(ctx context.Context, scalableDimension types.ScalableDimension, tableName string, minCapacity int32, maxCapacity int32) *ScalingError {
	request := applicationautoscaling.RegisterScalableTargetInput{
		MinCapacity:       &minCapacity,
		MaxCapacity:       &maxCapacity,
		ResourceId:        &tableName,
		ServiceNamespace:  DynamodbServiceNamespace,
		ScalableDimension: scalableDimension,
	}
	_, err := applicationAutoscalingClient.RegisterScalableTarget(ctx, &request)
	if err != nil {
		return &ScalingError{
			ServiceName:  string(DynamoDB),
			IdentifierId: tableName,
			Err:          err,
		}
	}
	return nil
}
func validateDynamoDBScalingConfig(clientConfig config.DynamoDBServiceScalingConfig) *ScalingError {
	if clientConfig.TableName == "" || validateRCUConfig(clientConfig.RCU) || validateWCUConfig(clientConfig.WCU) {
		return &ScalingError{
			ServiceName:  string(DynamoDB),
			IdentifierId: clientConfig.TableName,
			Err:          fmt.Errorf("invalid scaling config"),
		}
	}
	return nil
}

func validateRCUConfig(rcu config.RCU) bool {
	return rcu.MinProvisionedCapacity < 0 || rcu.MaxProvisionedCapacity < 0
}

func validateWCUConfig(wcu config.WCU) bool {
	return wcu.MinProvisionedCapacity < 0 || wcu.MaxProvisionedCapacity < 0
}
