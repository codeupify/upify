package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/infra"
	"github.com/codeupify/upify/internal/lang"
	"github.com/codeupify/upify/internal/platform/aws"
	"github.com/codeupify/upify/internal/platform/gcp"
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

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Add AWS configuration",
	RunE:  addAws,
}

var awsRegion string
var awsRuntime string

var gcpRegion string
var gcpProjectId string
var gcpRuntime string

var gcpCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Add GCP configuration",
	RunE:  addGCP,
}

func init() {
	rootCmd.AddCommand(platformCmd)
	platformCmd.AddCommand(platformAddCmd)
	platformCmd.AddCommand(platformListCmd)

	platformAddCmd.AddCommand(awsCmd)
	awsCmd.Flags().StringVar(&awsRegion, "region", "", "AWS region")
	awsCmd.Flags().StringVar(&awsRuntime, "runtime", "", "Lambda runtime")

	platformAddCmd.AddCommand(gcpCmd)
	gcpCmd.Flags().StringVar(&gcpRegion, "region", "", "GCP region")
	gcpCmd.Flags().StringVar(&gcpProjectId, "project-id", "", "GCP project ID")
	gcpCmd.Flags().StringVar(&gcpRuntime, "runtime", "", "Cloud Run runtime")
}

func addAws(cmd *cobra.Command, args []string) error {
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
		runtimes := getAWSLambdaRuntimes(cfg.Language)
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

	if err := aws.AddPlatform(cfg, awsRegion, awsRuntime); err != nil {
		return err
	}

	fmt.Println("Added AWS platform.")
	return nil
}

func addGCP(cmd *cobra.Command, args []string) error {
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
		runtimes := getGCPCloudRunRuntimes(cfg.Language)
		if len(runtimes) == 0 {
			return fmt.Errorf("CLI doesn't support the specified language yet for GCP: %s", cfg.Language)
		}

		runtimeQ := &survey.Select{
			Message: "Choose a runtime:",
			Options: runtimes,
		}
		if err := survey.AskOne(runtimeQ, &gcpRuntime); err != nil {
			return err
		}
	}

	if err := gcp.AddPlatform(cfg, gcpRegion, gcpRuntime, gcpProjectId); err != nil {
		return err
	}

	return nil
}

func listPlatforms() error {

	platforms := infra.ListPlatforms()
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

func getAWSLambdaRuntimes(language lang.Language) []string {
	switch language {
	case lang.Python:
		return []string{"python3.8", "python3.9", "python3.10", "python3.11", "python3.12"}
	case lang.JavaScript, lang.TypeScript:
		return []string{"nodejs18.x", "nodejs20.x"}
	default:
		return []string{}
	}
}

func getGCPCloudRunRuntimes(language lang.Language) []string {
	switch language {
	case lang.Python:
		return []string{"python37", "python38", "python39", "python310", "python311", "python312"}
	case lang.JavaScript, lang.TypeScript:
		return []string{"nodejs10", "nodejs12", "nodejs14", "nodejs16", "nodejs18", "nodejs20", "nodejs22"}
	default:
		return []string{}
	}
}
