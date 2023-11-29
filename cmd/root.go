package cmd

import (
	"fmt"
	"github.com/Cool-fire/aws-infra-scaler/pkg"
	"github.com/spf13/cobra"
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
	Use:   "ais",
	Short: "AWS Auto Scaler CLI",
	Long:  "AWS Auto Scaler CLI",
	Run: func(cmd *cobra.Command, args []string) {
		err := pkg.ScaleApp(options.scaleUpFlag, options.configPath)
		if err != nil {
			panic(err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
