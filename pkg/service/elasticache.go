package service

import (
	"context"
	"github.com/Cool-fire/aws-infra-scaler/pkg/config"
	"github.com/Cool-fire/aws-infra-scaler/pkg/errors"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
)

type ElasticCacheEngine string

const (
	Redis     ElasticCacheEngine = "redis"
	Memcached ElasticCacheEngine = "memcached"

	Other ElasticCacheEngine = "other"
)

type ElasticCacheService struct {
	Region string
	Client elasticache.Client
}

func (e ElasticCacheService) ScaleService(ctx context.Context, c config.ElasticCacheServiceScalingConfig, isScalingUp bool) *errors.ScalingFailureError {
	err := validateElasticCacheScalingConfig(c, isScalingUp, getElasticCacheEngine(c.Engine))
	if err != nil {
		return err
	}

	switch c.Engine {
	case "redis":
		return scaleRedis(ctx, c, isScalingUp, e.Client)
	case "memcached":
		return scaleMemcached(ctx, c, isScalingUp, e.Client)
	}

	return nil
}

func scaleRedis(ctx context.Context, clientConfig config.ElasticCacheServiceScalingConfig, up bool, client elasticache.Client) *errors.ScalingFailureError {

	nodeCount := int32(clientConfig.NodeCount)
	applyImmediately := true
	input := elasticache.ModifyReplicationGroupShardConfigurationInput{
		ApplyImmediately:   &applyImmediately,
		ReplicationGroupId: &clientConfig.ClusterId,
		NodeGroupCount:     &nodeCount,
	}

	if !up {
		input.NodeGroupsToRemove = clientConfig.NodesToDelete
	}

	_, err := client.ModifyReplicationGroupShardConfiguration(ctx, &input)
	return &errors.ScalingFailureError{
		ServiceName:  ElasticCache,
		IdentifierId: clientConfig.ClusterId,
		Reason:       err.Error(),
	}
}

func scaleMemcached(ctx context.Context, clientConfig config.ElasticCacheServiceScalingConfig, up bool, client elasticache.Client) *errors.ScalingFailureError {
	nodeCount := int32(clientConfig.NodeCount)
	applyImmediately := true
	input := elasticache.ModifyCacheClusterInput{
		ApplyImmediately: &applyImmediately,
		CacheClusterId:   &clientConfig.ClusterId,
		NumCacheNodes:    &nodeCount,
	}

	if !up {
		input.CacheNodeIdsToRemove = clientConfig.NodesToDelete
	}

	_, err := client.ModifyCacheCluster(ctx, &input)
	return &errors.ScalingFailureError{
		ServiceName:  ElasticCache,
		IdentifierId: clientConfig.ClusterId,
		Reason:       err.Error(),
	}
}

func getElasticCacheEngine(s string) ElasticCacheEngine {
	switch s {
	case "redis":
		return Redis
	case "memcached":
		return Memcached
	default:
		return Other
	}
}

func validateElasticCacheScalingConfig(clientConfig config.ElasticCacheServiceScalingConfig, isScalingUp bool, engine ElasticCacheEngine) *errors.ScalingFailureError {
	err := &errors.ScalingFailureError{
		ServiceName:  ElasticCache,
		IdentifierId: clientConfig.ClusterId,
		Reason:       "Invalid scaling config",
	}
	if engine == Other || clientConfig.ClusterId == "" || clientConfig.NodeCount <= 0 {
		return err
	}

	if !isScalingUp && len(clientConfig.NodesToDelete) == 0 {
		return err
	}
	return nil
}
