package config

import (
	"fmt"
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
