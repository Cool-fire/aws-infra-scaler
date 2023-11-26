package scaler

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
)

type ScalingConfig struct {
	Name           string          `yaml:"name"`
	AssumedRoleArn string          `yaml:"assumedRoleArn"`
	ScalingRegions []ScalingRegion `yaml:"scalingRegions"`
}

type ScalingRegion struct {
	Region              string        `yaml:"region"`
	ServiceScaleConfigs []interface{} `yaml:"serviceScaleConfigs"`
}

func (s *ScalingRegion) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var rawScalingRegion map[string]interface{}
	if err := unmarshal(&rawScalingRegion); err != nil {
		return err
	}

	for key, value := range rawScalingRegion {
		switch key {
		case "region":
			region, ok := value.(string)
			if !ok || region == "" {
				return fmt.Errorf("config error: Region field is missing or string is empty")
			}

			s.Region = region

		case "serviceScaleConfigs":
			v, ok := value.([]interface{})
			if !ok {
				return fmt.Errorf("config error: serviceScaleConfigs field is missing or not an array")
			}
			for _, serviceScaleConfig := range v {
				v, ok := serviceScaleConfig.(map[string]interface{})
				if !ok {
					return fmt.Errorf("config error: serviceScaleConfigs field is missing or not an array")
				}

				serviceConfig, err := convertMapToConfig(v)
				if err != nil {
					return err
				}

				s.ServiceScaleConfigs = append(s.ServiceScaleConfigs, serviceConfig)
			}
		}
	}

	return nil
}

type ServiceScalingConfig interface {
	GetName() string
}

type KinesisServiceScalingConfig struct {
	Service           string `mapstructure:"service"`
	StreamArn         string `mapstructure:"streamArn"`
	DesiredShardCount int    `mapstructure:"desiredShardCount"`
}

func (k KinesisServiceScalingConfig) GetName() string {
	return fmt.Sprintf("Kinesis scaling config for stream %s", k.StreamArn)
}

type EC2ServiceScalingConfig struct {
	Service      string `mapstructure:"service"`
	AsgName      string `mapstructure:"asgName"`
	MinCount     int    `mapstructure:"minCount"`
	DesiredCount int    `mapstructure:"desiredCount"`
	MaxCount     int    `mapstructure:"maxCount"`
}

func (e EC2ServiceScalingConfig) GetName() string {
	return fmt.Sprintf("EC2 scaling config for ASG %s", e.AsgName)
}

type ElasticCacheServiceScalingConfig struct {
	Service       string   `mapstructure:"service"`
	ClusterId     string   `mapstructure:"clusterId"`
	Engine        string   `mapstructure:"engine"`
	NodeCount     int      `mapstructure:"nodeCount"`
	NodesToDelete []string `mapstructure:"nodesToDelete"`
}

func (ec ElasticCacheServiceScalingConfig) GetName() string {
	return fmt.Sprintf("ElasticCache scaling config for cluster %s", ec.ClusterId)
}

type DynamoDBServiceScalingConfig struct {
	Service   string `mapstructure:"service"`
	TableName string `mapstructure:"tableName"`
	IsIndex   bool   `mapstructure:"isIndex"`
	RCU       RCU    `mapstructure:"rcu"`
	WCU       WCU    `mapstructure:"wcu"`
}

func (d DynamoDBServiceScalingConfig) GetName() string {
	return fmt.Sprintf("DynamoDB scaling config for table %s", d.TableName)
}

type RCU struct {
	MinProvisionedCapacity int `mapstructure:"minProvisionedCapacity"`
	MaxProvisionedCapacity int `mapstructure:"maxProvisionedCapacity"`
}

type WCU struct {
	MinProvisionedCapacity int `mapstructure:"minProvisionedCapacity"`
	MaxProvisionedCapacity int `mapstructure:"maxProvisionedCapacity"`
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
