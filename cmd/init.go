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
		return false, fmt.Errorf("failed to read package.json: %w", err)
	}
	var packageJSON map[string]interface{}
	if err := json.Unmarshal(packageContentBytes, &packageJSON); err != nil {
		return false, fmt.Errorf("failed to parse package.json: %w", err)
	}

	return packageJSON["type"] == "module", nil
}

func determineModuleSystem(entrypointPath string) (config.ModuleSystem, error) {

	es6, err := detectES6FromPackageJSON()
	if err == nil && es6 {
		return config.ES6, nil
	}

	if entrypointPath != "" {
		contentBytes, err := os.ReadFile(entrypointPath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		content := string(contentBytes)

		isCommonJS := strings.Contains(content, "require(") || strings.Contains(content, "module.exports")
		isES6 := strings.Contains(content, "import ") || strings.Contains(content, "export ")

		if isCommonJS {
			return config.CommonJS, nil
		} else if isES6 {
			return config.ES6, nil
		}
	}

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

	var moduleSystem string
	if err := survey.Ask(moduleSystemQ, &moduleSystem); err != nil {
		return "", fmt.Errorf("failed to get module system: %w", err)
	}

	return config.ModuleSystem(moduleSystem), nil
}

func determinePackageManager(language config.Language) (config.PackageManager, error) {
	switch language {
	case config.Python:
		return config.Pip, nil
	case config.JavaScript, config.TypeScript:
		if _, err := os.Stat("yarn.lock"); err == nil {
			return config.Yarn, nil
		}
		if _, err := os.Stat("package-lock.json"); err == nil {
			return config.Npm, nil
		}
		return config.Npm, nil
	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}
}

func determineName(providedName string) (string, error) {
	if providedName != "" {
		return providedName, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	return filepath.Base(cwd), nil
}

func determineLanguage(entrypoint string) (config.Language, error) {
	ext := filepath.Ext(entrypoint)
	switch ext {
	case ".py":
		return config.Python, nil
	case ".js":
		return config.JavaScript, nil
	case ".ts":
		return config.TypeScript, nil
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes Upify",
	Long:  "Creates the .upify folder and a basic config.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			selectedFramework config.Framework
			selectedLanguage  config.Language
			moduleSystem      config.ModuleSystem
			entrypoint        string
			appVar            string
		)

		frameworkQ := []*survey.Question{
			{
				Name:     "framework",
				Prompt:   &survey.Select{Message: "Choose a framework:", Options: []string{"express", "flask", "other/none"}},
				Validate: survey.Required,
			},
		}

		var framework string
		if err := survey.Ask(frameworkQ, &framework); err != nil {
			return err
		}

		switch framework {
		case "flask":
			entrypointPrompt := &survey.Input{
				Message: "Enter the relative path to your main application file (e.g., main.py):",
				Default: "main.py",
			}
			if err := survey.AskOne(entrypointPrompt, &entrypoint, survey.WithValidator(survey.Required)); err != nil {
				return err
			}

			if err := validateEntrypoint(entrypoint); err != nil {
				return err
			}

			appVarPrompt := &survey.Input{
				Message: "Enter the name of the Flask application instance variable (default 'app'):",
				Default: "app",
			}
			if err := survey.AskOne(appVarPrompt, &appVar, survey.WithValidator(survey.Required)); err != nil {
				return err
			}

			selectedFramework = config.Framework(framework)
			selectedLanguage = config.Python

		case "express":

			entrypointPrompt := &survey.Input{
				Message: "Enter the relative path to your main application file (e.g., app.js):",
				Default: "app.js",
			}
			if err := survey.AskOne(entrypointPrompt, &entrypoint, survey.WithValidator(survey.Required)); err != nil {
				return err
			}

			if err := validateEntrypoint(entrypoint); err != nil {
				return err
			}

			appVarPrompt := &survey.Input{
				Message: "Enter the name of the Express export variable (default 'app'):",
				Default: "app",
			}
			if err := survey.AskOne(appVarPrompt, &appVar, survey.WithValidator(survey.Required)); err != nil {
				return err
			}

			language, err := determineLanguage(entrypoint)
			if err != nil {
				return err
			}

			selectedFramework = config.Framework(framework)
			selectedLanguage = language

		case "other/none":
			languageQ := []*survey.Question{
				{
					Name:     "language",
					Prompt:   &survey.Select{Message: "Choose the language/runtime:", Options: []string{"python", "javascript", "typescript"}},
					Validate: survey.Required,
				},
			}

			var language string
			if err := survey.Ask(languageQ, &language); err != nil {
				return err
			}

			selectedLanguage = config.Language(language)
		}

		if selectedLanguage == config.JavaScript || selectedLanguage == config.TypeScript {
			ms, err := determineModuleSystem(entrypoint)
			if err != nil {
				return err
			}

			moduleSystem = ms
		}

		projectNameQ := []*survey.Question{
			{
				Name:   "projectName",
				Prompt: &survey.Input{Message: "Enter a custom project name (leave blank to use current directory name):"},
			},
		}

		var projectName string
		if err := survey.Ask(projectNameQ, &projectName); err != nil {
			return err
		}

		projectName, err := determineName(projectName)
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

		packageManager, err := determinePackageManager(selectedLanguage)
		if err != nil {
			return err
		}

		cfg := &config.Config{
			Framework:      selectedFramework,
			Language:       selectedLanguage,
			PackageManager: packageManager,
			Entrypoint:     entrypoint,
			Name:           projectName,
			AppVar:         appVar,
			ModuleSystem:   moduleSystem,
		}

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Println("Configuration saved successfully.")
		return nil
	},
}
