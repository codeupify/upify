package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/platforms/aws/lambda"
	"github.com/spf13/cobra"
)

var platformCmd = &cobra.Command{
	Use:   "platform",
	Short: "Manage platform configurations",
}

var platformAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new platform configuration",
}

var platformListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured platforms",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listPlatforms()
	},
}

var awsLambdaCmd = &cobra.Command{
	Use:   "aws-lambda",
	Short: "Add AWS Lambda configuration",
	RunE:  addAWSLambda,
}

var awsRegion string
var awsRuntime string

var gcpCloudRunCmd = &cobra.Command{
	Use:   "gcp-cloudrun",
	Short: "Add GCP Cloud Run configuration",
	RunE:  addGCPCloudRun,
}

func init() {
	rootCmd.AddCommand(platformCmd)
	platformCmd.AddCommand(platformAddCmd)
	platformCmd.AddCommand(platformListCmd)

	platformAddCmd.AddCommand(awsLambdaCmd)
	awsLambdaCmd.Flags().StringVar(&awsRegion, "region", "", "AWS region")
	awsLambdaCmd.Flags().StringVar(&awsRuntime, "runtime", "", "Lambda runtime")

	platformAddCmd.AddCommand(gcpCloudRunCmd)
}

func getAWSRuntimes(language config.Language) []string {
	switch language {
	case config.Python:
		return []string{"python3.8", "python3.9", "python3.10", "python3.11", "python3.12"}
	case config.JavaScript, config.TypeScript:
		return []string{"nodejs18.x", "nodejs20.x"}
	default:
		return []string{}
	}
}

func addAWSLambda(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if awsRegion == "" {
		regionQ := &survey.Input{
			Message: "Enter AWS region:",
			Default: "us-east-1",
		}
		if err := survey.AskOne(regionQ, &awsRegion); err != nil {
			return err
		}
	}

	if awsRuntime == "" {
		runtimes := getAWSRuntimes(cfg.Language)
		if len(runtimes) == 0 {
			return fmt.Errorf("CLI doesn't support the specified language yet for AWS Lambda: %s", cfg.Language)
		}

		runtimeQ := &survey.Select{
			Message: "Choose a runtime:",
			Options: runtimes,
		}
		if err := survey.AskOne(runtimeQ, &awsRuntime); err != nil {
			return err
		}
	}

	if err := lambda.AddConfig(cfg, awsRegion, awsRuntime); err != nil {
		return err
	}

	if err := lambda.GenerateLambdaHandler(cfg); err != nil {
		return err
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Added AWS Lambda configuration successfully.")
	return nil
}

func addGCPCloudRun(cmd *cobra.Command, args []string) error {
	fmt.Println("GCP Cloud Run configuration not yet implemented.")
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
