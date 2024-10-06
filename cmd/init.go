package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/codeupify/upify/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

func validateEntrypoint(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}
	return nil
}

func detectES6FromPackageJSON() (bool, error) {
	packageJSONPath := "./package.json"
	packageContentBytes, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return false, nil
	}
	var packageJSON map[string]interface{}
	if err := json.Unmarshal(packageContentBytes, &packageJSON); err != nil {
		return false, fmt.Errorf("failed to parse package.json: %w", err)
	}

	return packageJSON["type"] == "module", nil
}

func detectModuleSystem(entrypointPath string) (string, error) {
	es6, err := detectES6FromPackageJSON()
	if err == nil && es6 {
		return "es6", nil
	}

	contentBytes, err := os.ReadFile(entrypointPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	content := string(contentBytes)

	isCommonJS := strings.Contains(content, "require(") || strings.Contains(content, "module.exports")
	isES6 := strings.Contains(content, "import ") || strings.Contains(content, "export ")

	if isCommonJS {
		return "commonjs", nil
	} else if isES6 {
		return "es6", nil
	}

	return "", fmt.Errorf("could not determine module system")
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

func determineName(providedName string) (string, error) {
	if providedName != "" {
		return providedName, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		var userInput string
		prompt := &survey.Input{
			Message: "Couldn't determine the current directory name. Please enter a project name:",
		}
		err := survey.AskOne(prompt, &userInput, survey.WithValidator(survey.Required))
		if err != nil {
			return "", fmt.Errorf("failed to get project name: %w", err)
		}
		return userInput, nil
	}
	return filepath.Base(cwd), nil
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

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes Upify",
	Long:  "Creates the .upify folder and a basic config.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		answers := struct {
			Framework    string
			Entrypoint   string
			AppVar       string
			Language     string
			ModuleSystem string
			ProjectName  string
		}{}

		frameworkQ := []*survey.Question{
			{
				Name:     "framework",
				Prompt:   &survey.Select{Message: "Choose a framework:", Options: []string{"express", "flask", "none"}},
				Validate: survey.Required,
			},
		}

		if err := survey.Ask(frameworkQ, &answers.Framework); err != nil {
			return err
		}

		switch answers.Framework {
		case "flask":
			entrypointPrompt := &survey.Input{
				Message: "Enter the relative path to your main application file (e.g., main.py):",
				Default: "main.py",
			}
			if err := survey.AskOne(entrypointPrompt, &answers.Entrypoint, survey.WithValidator(survey.Required)); err != nil {
				return err
			}

			if err := validateEntrypoint(answers.Entrypoint); err != nil {
				return err
			}

			appVarPrompt := &survey.Input{
				Message: "Enter the name of the Flask application instance variable (default 'app'):",
				Default: "app",
			}
			if err := survey.AskOne(appVarPrompt, &answers.AppVar, survey.WithValidator(survey.Required)); err != nil {
				return err
			}

			answers.Language = "python"

		case "express":

			entrypointPrompt := &survey.Input{
				Message: "Enter the relative path to your main application file (e.g., app.js):",
				Default: "app.js",
			}
			if err := survey.AskOne(entrypointPrompt, &answers.Entrypoint, survey.WithValidator(survey.Required)); err != nil {
				return err
			}

			if err := validateEntrypoint(answers.Entrypoint); err != nil {
				return err
			}

			appVarPrompt := &survey.Input{
				Message: "Enter the name of the Express export variable (default 'app'):",
				Default: "app",
			}
			if err := survey.AskOne(appVarPrompt, &answers.AppVar, survey.WithValidator(survey.Required)); err != nil {
				return err
			}

			detected, err := detectModuleSystem(answers.Entrypoint)
			if err != nil {
				fmt.Println("For a JavaScript project, you must choose the module system. CommonJS uses require() and ES6 uses import/export.")
				moduleSystemQ := []*survey.Question{
					{
						Name: "moduleSystem",
						Prompt: &survey.Select{
							Message: "Choose the module system (defaults to commonjs):",
							Options: []string{"commonjs", "es6"},
							Default: "commonjs",
						},
						Validate: survey.Required,
					},
				}

				if err := survey.Ask(moduleSystemQ, &answers.ModuleSystem); err != nil {
					return err
				}
			} else {
				answers.ModuleSystem = detected
				fmt.Printf("Auto-detected module system: %s\n", answers.ModuleSystem)
			}

			answers.Language = determineLanguage(answers.Entrypoint)

		case "none":
			languageQ := []*survey.Question{
				{
					Name:     "language",
					Prompt:   &survey.Select{Message: "Choose the language/runtime:", Options: []string{"python", "javascript", "typescript"}},
					Validate: survey.Required,
				},
			}

			if err := survey.Ask(languageQ, &answers.Language); err != nil {
				return err
			}

			if answers.Language == "javascript" || answers.Language == "typescript" {
				es6, err := detectES6FromPackageJSON()
				if err != nil {
					return err
				}
				if es6 {
					answers.ModuleSystem = "es6"
				} else {
					fmt.Println("For a JavaScript project, you must choose the module system. CommonJS uses require() and ES6 uses import/export.")
					moduleSystemQ := &survey.Select{
						Message: "Choose the module system (defaults to commonjs):",
						Options: []string{"commonjs", "es6"},
						Default: "commonjs",
					}

					if err := survey.AskOne(moduleSystemQ, &answers.ModuleSystem); err != nil {
						return err
					}
				}
			}

		}

		projectNameQ := []*survey.Question{
			{
				Name:   "projectName",
				Prompt: &survey.Input{Message: "Enter a custom project name (leave blank to use current directory name):"},
			},
		}

		if err := survey.Ask(projectNameQ, &answers.ProjectName); err != nil {
			return err
		}

		name, err := determineName(answers.ProjectName)
		if err != nil {
			return err
		}

		if config.ConfigExists() {
			confirmOverwriteQ := []*survey.Question{
				{
					Name:     "confirmOverwrite",
					Prompt:   &survey.Confirm{Message: "Existing configuration found. Overwrite?", Default: false},
					Validate: survey.Required,
				},
			}

			var confirmOverwrite bool
			if err := survey.Ask(confirmOverwriteQ, &confirmOverwrite); err != nil {
				return err
			}

			if !confirmOverwrite {
				fmt.Println("Configuration not overwritten.")
				return nil
			}
		}

		cfg := &config.Config{
			Framework:      answers.Framework,
			Language:       answers.Language,
			PackageManager: determinePackageManager(answers.Language),
			Entrypoint:     answers.Entrypoint,
			Name:           name,
			AppVar:         answers.AppVar,
			ModuleSystem:   answers.ModuleSystem,
		}

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Println("Configuration saved successfully.")
		return nil
	},
}
