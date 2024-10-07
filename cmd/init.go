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

func promptModuleSystem() (config.ModuleSystem, error) {
	moduleSystemQ := []*survey.Question{
		{
			Name: "moduleSystem",
			Prompt: &survey.Select{
				Message: "Choose the module system:",
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
	if err != nil {
		fmt.Printf("Warning: %v. Couldn't determine module system from package.json.\n", err)
	}

	if es6 {
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

	return promptModuleSystem()
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

func askFramework() (string, error) {
	frameworkQ := []*survey.Question{
		{
			Name:     "framework",
			Prompt:   &survey.Select{Message: "Choose a framework:", Options: []string{"express", "flask", "other/none"}},
			Validate: survey.Required,
		},
	}

	var framework string
	if err := survey.Ask(frameworkQ, &framework); err != nil {
		return "", err
	}

	return framework, nil
}

func askLanguage() (config.Language, error) {
	languageQ := []*survey.Question{
		{
			Name:     "language",
			Prompt:   &survey.Select{Message: "Choose the language/runtime:", Options: []string{"python", "javascript", "typescript"}},
			Validate: survey.Required,
		},
	}

	var language string
	if err := survey.Ask(languageQ, &language); err != nil {
		return "", err
	}

	return config.Language(language), nil
}

func askEntrypoint(message string, defaultValue string) (string, error) {
	entrypointPrompt := &survey.Input{
		Message: message,
		Default: defaultValue,
	}

	var entrypoint string
	if err := survey.AskOne(entrypointPrompt, &entrypoint, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	if err := validateEntrypoint(entrypoint); err != nil {
		return "", err
	}

	return entrypoint, nil
}

func askAppVar(message string, defaultValue string) (string, error) {
	appVarPrompt := &survey.Input{
		Message: message,
		Default: defaultValue,
	}

	var appVar string
	if err := survey.AskOne(appVarPrompt, &appVar, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	return appVar, nil
}

func askProjectName() (string, error) {
	projectNameQ := []*survey.Question{
		{
			Name:   "projectName",
			Prompt: &survey.Input{Message: "Enter a custom project name (leave blank to use current directory name):"},
		},
	}

	var projectName string
	if err := survey.Ask(projectNameQ, &projectName); err != nil {
		return "", err
	}

	projectName, err := determineName(projectName)
	if err != nil {
		return "", err
	}

	return projectName, nil
}

func askOverwriteConfirmation() (bool, error) {
	confirmOverwriteQ := []*survey.Question{
		{
			Name:     "confirmOverwrite",
			Prompt:   &survey.Confirm{Message: "Existing configuration found. Overwrite?", Default: true},
			Validate: survey.Required,
		},
	}

	var confirmOverwrite bool
	if err := survey.Ask(confirmOverwriteQ, &confirmOverwrite); err != nil {
		return false, err
	}

	return confirmOverwrite, nil
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
			projectName       string
		)

		framework, err := askFramework()
		if err != nil {
			return err
		}

		switch framework {
		case "flask":

			ret, err := askEntrypoint("Enter the relative path to your main application file (e.g., app.py):", "app.py")
			if err != nil {
				return err
			}
			entrypoint = ret

			ret, err = askAppVar("Enter the name of the Flask application instance variable (default 'app'):", "app")
			if err != nil {
				return err
			}
			appVar = ret

			selectedFramework = config.Framework(framework)
			selectedLanguage = config.Python

		case "express":

			ret, err := askEntrypoint("Enter the relative path to your main application file (e.g., app.js):", "app.js")
			if err != nil {
				return err
			}
			entrypoint = ret

			ret, err = askAppVar("Enter the name of the Express application instance variable (default 'app'):", "app")
			if err != nil {
				return err
			}
			appVar = ret

			language, err := determineLanguage(entrypoint)
			if err != nil {
				return err
			}

			selectedFramework = config.Framework(framework)
			selectedLanguage = language

		case "other/none":

			language, err := askLanguage()
			if err != nil {
				return err
			}

			selectedLanguage = language
		}

		if selectedLanguage == config.JavaScript || selectedLanguage == config.TypeScript {
			ms, err := determineModuleSystem(entrypoint)
			if err != nil {
				return err
			}

			moduleSystem = ms
		}

		ret, err := askProjectName()
		if err != nil {
			return err
		}

		projectName = ret

		if config.ConfigExists() {
			confirmOverwrite, err := askOverwriteConfirmation()
			if err != nil {
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
