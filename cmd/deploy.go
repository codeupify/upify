package cmd

import (
	"fmt"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/platform"
	"github.com/codeupify/upify/internal/platform/aws"
	"github.com/codeupify/upify/internal/platform/gcp"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [platform]",
	Short: "Deploy the application to a specified platform",
	Long: `Deploy the application to a specified platform.
Currently supported platforms: aws, gcp

Example:
  upify deploy aws`,
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

func deploy(platformStr string, cfg *config.Config) error {
	switch platformStr {
	case string(platform.AWS):
		fmt.Println("Deploying to AWS...")
		if err := aws.Deploy(cfg); err != nil {
			return fmt.Errorf("failed to deploy to AWS: %w", err)
		}
	case string(platform.GCP):
		fmt.Println("Deploying to GCP...")
		if err := gcp.Deploy(cfg); err != nil {
			return fmt.Errorf("failed to deploy to GCP: %w", err)
		}
	default:
		return fmt.Errorf("unsupported platform: %s", platformStr)
	}

	return nil
}
