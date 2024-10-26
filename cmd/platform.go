package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/lang"
	"github.com/codeupify/upify/internal/platform/aws/lambda"
	"github.com/codeupify/upify/internal/platform/gcp/cloudrun"
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

var gcpRegion string
var gcpProjectId string
var gcpRuntime string

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
	gcpCloudRunCmd.Flags().StringVar(&gcpRegion, "region", "", "GCP region")
	gcpCloudRunCmd.Flags().StringVar(&gcpProjectId, "project-id", "", "GCP project ID")
	gcpCloudRunCmd.Flags().StringVar(&gcpRuntime, "runtime", "", "Cloud Run runtime")
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

	if err := lambda.AddHandler(cfg); err != nil {
		return err
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Added AWS Lambda platform.")
	return nil
}

func addGCPCloudRun(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if gcpRegion == "" {
		regionQ := &survey.Input{
			Message: "Enter GCP region:",
			Default: "us-central1",
		}
		if err := survey.AskOne(regionQ, &gcpRegion); err != nil {
			return err
		}
	}

	if gcpProjectId == "" {
		projectIdQ := &survey.Input{
			Message: "Enter GCP project ID:",
		}
		if err := survey.AskOne(projectIdQ, &gcpProjectId); err != nil {
			return err
		}
	}

	if gcpRuntime == "" {
		runtimes := getGCPRuntimes(cfg.Language)
		if len(runtimes) == 0 {
			return fmt.Errorf("CLI doesn't support the specified language yet for GCP Cloud Run: %s", cfg.Language)
		}

		runtimeQ := &survey.Select{
			Message: "Choose a runtime:",
			Options: runtimes,
		}
		if err := survey.AskOne(runtimeQ, &gcpRuntime); err != nil {
			return err
		}
	}

	if err := cloudrun.AddConfig(cfg, gcpRegion, gcpProjectId, gcpRuntime); err != nil {
		return err
	}

	if err := cloudrun.AddHandler(cfg); err != nil {
		return err
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

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

func getAWSRuntimes(language lang.Language) []string {
	switch language {
	case lang.Python:
		return []string{"python3.8", "python3.9", "python3.10", "python3.11", "python3.12"}
	case lang.JavaScript, lang.TypeScript:
		return []string{"nodejs18.x", "nodejs20.x"}
	default:
		return []string{}
	}
}

func getGCPRuntimes(language lang.Language) []string {
	switch language {
	case lang.Python:
		return []string{"python37", "python38", "python39", "python310", "python311", "python312"}
	case lang.JavaScript, lang.TypeScript:
		return []string{"nodejs10", "nodejs12", "nodejs14", "nodejs16", "nodejs18", "nodejs20", "nodejs22"}
	default:
		return []string{}
	}
}
