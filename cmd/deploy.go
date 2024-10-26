package cmd

import (
	"fmt"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/platform/aws/lambda"
	"github.com/codeupify/upify/internal/platform/gcp/cloudrun"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [platform]",
	Short: "Deploy the application to a specified platform",
	Long: `Deploy the application to a specified platform.
Currently supported platforms: aws-lambda, gcp-cloudrun

Example:
  upify deploy aws-lambda`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		platform := args[0]
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		return deploy(platform, cfg)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func deploy(platform string, cfg *config.Config) error {
	switch platform {
	case "aws-lambda":
		if cfg.AWSLambda == nil {
			return fmt.Errorf("aws-lambda configuration is not set up. Please run 'upify platform add aws-lambda' first")
		}
		fmt.Println("Deploying to AWS Lambda...")
		if err := lambda.Deploy(cfg); err != nil {
			return fmt.Errorf("failed to deploy to AWS Lambda: %w", err)
		}
	case "gcp-cloudrun":
		if cfg.GCPCloudRun == nil {
			return fmt.Errorf("gcp-cloudrun configuration is not set up. Please run 'upify platform add gcp-cloudrun' first")
		}
		fmt.Println("Deploying to GCP Cloud Run...")
		if err := cloudrun.Deploy(cfg); err != nil {
			return fmt.Errorf("failed to deploy to GCP Cloud Run: %w", err)
		}
	default:
		return fmt.Errorf("unsupported platform: %s", platform)
	}

	return nil
}
