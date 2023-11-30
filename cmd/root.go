package cmd

import (
	"fmt"
	"github.com/Cool-fire/aws-infra-scaler/pkg"
	"github.com/spf13/cobra"
	"log"
)

type Options struct {
	scaleUpFlag   bool
	scaleDownFlag bool
	configPath    string
}

var options *Options

func init() {
	options = &Options{}

	rootCmd.PersistentFlags().BoolVarP(&options.scaleUpFlag, "scale-up", "u", false, "Scale up")
	rootCmd.PersistentFlags().BoolVarP(&options.scaleDownFlag, "scale-down", "d", false, "Scale down")
	rootCmd.PersistentFlags().StringVarP(&options.configPath, "config-path", "c", "", "Config file path")

	// make config flag required
	_ = rootCmd.MarkPersistentFlagRequired("config")
}

var rootCmd = &cobra.Command{
	Use:   "aws-infra-scaler",
	Short: "AWS Auto Scaler CLI is a simple CLI tool to scale AWS infrastructure services via YAML config files",
	Long: `AWS Auto Scaler CLI is a CLI tool to scale AWS infrastructure services via YAML config files, 
			It is designed to scale AWS infrastructure services such as DynamoDB, Kinesis, Elasticache, EC2 etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		scalingResponse, err := pkg.ScaleApp(options.scaleUpFlag, options.configPath)
		if err != nil {
			log.Fatalf("error scaling app: %v", err)
		}

		if scalingResponse.ContainsFailedServices {
			log.Printf("scaling completed with errors")
			for region, scalingErrors := range scalingResponse.RegionalFailedServices {
				fmt.Printf("----------region: %s------------\n", region)
				for i, scalingError := range scalingErrors {
					if i != 0 {
						fmt.Println("------------------------------------------------")
					}
					fmt.Printf("service: %s\nidentifier: %s\nerror: %v\n", scalingError.ServiceName, scalingError.IdentifierId, scalingError.Err)
				}
			}
		} else {
			log.Printf("scaling completed successfully")
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
