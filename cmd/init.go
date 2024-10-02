package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/codeupify/upify/internal/config"
	"github.com/spf13/cobra"
)

var (
	framework  string
	entrypoint string
	overwrite  bool
	manualMode bool
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&framework, "framework", "f", "", "Framework used (none, flask, express)")
	initCmd.Flags().StringVarP(&entrypoint, "entrypoint", "e", "", "Path to your main application file")
	initCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "Overwrite existing files")
}

func getParams(cmd *cobra.Command) error {
	manualMode = cmd.Flags().NFlag() > 0

	if manualMode {
		if framework == "" || entrypoint == "" {
			fmt.Println("Error: When providing manual parameters, all of the following are required:")
			fmt.Println("  -f, --framework")
			fmt.Println("  -e, --entrypoint")
			return fmt.Errorf("missing required flags")
		}
	} else {
		questions := []*survey.Question{
			{
				Name:     "framework",
				Prompt:   &survey.Select{Message: "Choose a framework:", Options: []string{"none", "flask", "express"}},
				Validate: survey.Required,
			},
			{
				Name:   "entrypoint",
				Prompt: &survey.Input{Message: "Enter the path to your main application file:"},
				Validate: func(val interface{}) error {
					str, ok := val.(string)
					if !ok {
						return fmt.Errorf("invalid input")
					}
					if _, err := os.Stat(str); os.IsNotExist(err) {
						return fmt.Errorf("file does not exist: %s", str)
					}
					if !filepath.IsAbs(str) {
						absPath, err := filepath.Abs(str)
						if err == nil {
							fmt.Printf("Note: Using absolute path: %s\n", absPath)
						}
					}
					return nil
				},
			},
		}

		answers := struct {
			Framework  string
			Entrypoint string
		}{}

		err := survey.Ask(questions, &answers)
		if err != nil {
			return err
		}

		framework = answers.Framework
		entrypoint = answers.Entrypoint
	}

	return nil
}

func determineLanguage(entrypoint string) string {
	ext := filepath.Ext(entrypoint)
	switch ext {
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	default:
		return "unknown"
	}
}

func determinePackageManager(language string) string {
	switch language {
	case "python":
		return "pip"
	case "javascript", "typescript":
		if _, err := os.Stat("yarn.lock"); err == nil {
			return "yarn"
		}
		if _, err := os.Stat("package-lock.json"); err == nil {
			return "npm"
		}
		return "npm"
	default:
		return "unknown"
	}
}

func checkOverwrite() (bool, error) {
	if manualMode {
		return overwrite, nil
	}

	var confirmOverwrite bool
	prompt := &survey.Confirm{
		Message: "Existing configuration found. Overwrite?",
	}
	err := survey.AskOne(prompt, &confirmOverwrite)
	if err != nil {
		return false, fmt.Errorf("failed to get overwrite confirmation: %w", err)
	}

	return confirmOverwrite, nil
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize your project for deployment",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := getParams(cmd); err != nil {
			return err
		}

		language := determineLanguage(entrypoint)
		packageManager := determinePackageManager(language)

		if config.ConfigExists() {
			shouldOverwrite, err := checkOverwrite()
			if err != nil {
				return err
			}

			if !shouldOverwrite {
				return nil
			}
		}

		cfg := &config.Config{
			Framework:      framework,
			Language:       language,
			PackageManager: packageManager,
			Entrypoint:     entrypoint,
		}

		if err := config.SaveConfig(cfg); err != nil {
			return err
		}

		fmt.Println("Configuration saved successfully.")
		return nil
	},
}
