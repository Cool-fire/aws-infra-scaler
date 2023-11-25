package scaler

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func ScaleApplication(shouldScaleUp bool, configPath string) error {
	scalingConfig, err := readConfig(configPath)
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	fmt.Printf("Scaling config: %+v\n", *scalingConfig)
	return nil
}

func readConfig(configPath string) (*ScalingConfig, error) {
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
