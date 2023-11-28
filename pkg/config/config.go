package config

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
	"os"
)

func ReadConfig(configPath string) (*ScalingConfig, error) {
	fmt.Println("Reading config from path: ", configPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var scalingConfig ScalingConfig
	if err := yaml.Unmarshal(data, &scalingConfig); err != nil {
		return nil, fmt.Errorf("error decoding config file: %w", err)
	}

	return &scalingConfig, nil
}

func decodeConfig(decoderConfig *mapstructure.DecoderConfig, data map[string]interface{}) error {
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return fmt.Errorf("error creating decoder: %w", err)
	}

	err = decoder.Decode(data)
	if err != nil {
		return err
	}

	return nil
}

func convertMapToConfig(data map[string]interface{}) (interface{}, error) {
	s, ok := data["service"].(string)
	if !ok || s == "" {
		return nil, fmt.Errorf("config error: Service field is missing or string is empty")
	}

	decoderConfig := mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		ErrorUnused:      true,
		ErrorUnset:       true,
	}

	switch s {
	case "Kinesis":
		var kinesisServiceScalingConfig KinesisServiceScalingConfig
		decoderConfig.Result = &kinesisServiceScalingConfig

		err := decodeConfig(&decoderConfig, data)
		if err != nil {
			return nil, fmt.Errorf("error decoding Kinesis service scaling config: %w", err)
		}
		return kinesisServiceScalingConfig, nil

	case "Ec2":
		var ec2ServiceScalingConfig EC2ServiceScalingConfig
		decoderConfig.Result = &ec2ServiceScalingConfig

		err := decodeConfig(&decoderConfig, data)
		if err != nil {
			return nil, fmt.Errorf("error decoding Ec2 service scaling config: %w", err)
		}
		return ec2ServiceScalingConfig, nil

	case "ElasticCache":
		var elasticCacheServiceScalingConfig ElasticCacheServiceScalingConfig
		decoderConfig.Result = &elasticCacheServiceScalingConfig

		err := decodeConfig(&decoderConfig, data)
		if err != nil {
			return nil, fmt.Errorf("error decoding ElasticCache service scaling config: %w", err)
		}
		return elasticCacheServiceScalingConfig, nil

	case "DynamoDB":
		var dynamoDBServiceScalingConfig DynamoDBServiceScalingConfig
		decoderConfig.Result = &dynamoDBServiceScalingConfig

		err := decodeConfig(&decoderConfig, data)
		if err != nil {
			return nil, fmt.Errorf("error decoding DynamoDB service scaling config: %w", err)
		}
		return dynamoDBServiceScalingConfig, nil

	default:
		fmt.Println("Service is not supported")
		return nil, fmt.Errorf("config error: Service %s is not supported", s)
	}
}
