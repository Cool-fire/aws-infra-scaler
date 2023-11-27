package pkg

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

type ScalingResult struct {
	region  string
	service string
	err     error
}

func ScaleApplication(shouldScaleUp bool, configPath string) error {
	scalingConfig, err := readConfig(configPath)
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan int)

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

func scaleRegion(ctx context.Context, scalingRegion ScalingRegion, shouldScaleUp bool, wg *sync.WaitGroup, resultChan chan int) {
	defer wg.Done()

	var serviceWg sync.WaitGroup

	for _, serviceScaleConfig := range scalingRegion.ServiceScaleConfigs {
		serviceWg.Add(1)
		go scaleService(ctx, serviceScaleConfig, shouldScaleUp, &serviceWg, resultChan)
	}

	serviceWg.Wait()
	fmt.Println("done scaling region ", scalingRegion.Region)
}

func scaleService(ctx context.Context, serviceScaleConfig interface{}, shouldScaleUp bool, wg *sync.WaitGroup, resultChan chan int) {
	defer wg.Done()

	switch serviceScaleConfig.(type) {
	case KinesisServiceScalingConfig:
		fmt.Println("Scaling Kinesis")
		resultChan <- 1
	case EC2ServiceScalingConfig:
		fmt.Println("Scaling EC2")
		resultChan <- 2
	default:
		fmt.Println("Unknown service")
		resultChan <- 3
	}
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
