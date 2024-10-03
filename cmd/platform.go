package cmd

import (
	"fmt"
	"strings"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/platforms/aws/lambda"
	"github.com/spf13/cobra"
)

var supportedPlatforms = []string{"aws-lambda", "gcp-cloudrun"}

var platformCmd = &cobra.Command{
	Use:   "platform",
	Short: "Manage platform configurations",
}

var platformAddCmd = &cobra.Command{
	Use:   "add [platform]",
	Short: "Add a new platform configuration",
	Long: fmt.Sprintf(`Add a new platform configuration.
	
Available platforms: %s

Example:
  upify platform add aws-lambda`, strings.Join(supportedPlatforms, ", ")),
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		platform := args[0]
		return addPlatform(platform)
	},
}

var platformListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured platforms",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listPlatforms()
	},
}

func init() {
	rootCmd.AddCommand(platformCmd)
	platformCmd.AddCommand(platformAddCmd)
	platformCmd.AddCommand(platformListCmd)
}

func addPlatform(platform string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch platform {
	case "aws-lambda":
		if err := lambda.AddConfig(cfg); err != nil {
			return err
		}
		if err := lambda.GenerateLambdaHandler(cfg); err != nil {
			return err
		}
	case "gcp-cloudrun":
		if cfg.GCPCloudRun != nil {
			return fmt.Errorf("gcp-cloudrun configuration already exists")
		}
		cfg.GCPCloudRun = &config.GCPCloudRunConfig{
			// TODO
		}
	default:
		return fmt.Errorf("unsupported platform: %s", platform)
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Added %s configuration successfully.\n", platform)
	return nil
}

func listPlatforms() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var platforms []string

	if cfg.AWSLambda != nil {
		platforms = append(platforms, "aws-lambda")
	}
	if cfg.GCPCloudRun != nil {
		platforms = append(platforms, "gcp-cloudrun")
	}

	if len(platforms) == 0 {
		fmt.Println("No platforms configured.")
	} else {
		fmt.Println("Configured platforms:")
		for _, platform := range platforms {
			fmt.Printf("- %s\n", platform)
		}
	}

	return nil
}
