package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/framework"
	"github.com/codeupify/upify/internal/lang"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes Upify",
	Long:  "Creates the .upify folder and a basic config.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			selectedFramework framework.Framework
			selectedLanguage  lang.Language
			entrypoint        string
			appVar            string
			projectName       string
		)

		frameworkStr, err := askFramework()
		if err != nil {
			return err
		}

		switch frameworkStr {
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

			selectedFramework = framework.Framework(frameworkStr)
			selectedLanguage = lang.Python

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

			selectedFramework = framework.Framework(frameworkStr)
			selectedLanguage = language

		case "other/none":

			language, err := askLanguage()
			if err != nil {
				return err
			}

			selectedLanguage = language
		}

		// if selectedLanguage == config.Python {
		// 	pythonVersion, err := detectPythonVersion()
		// 	if err != nil {
		// 		return err
		// 	}
		// 	ret, err := askRuntimeVersion(pythonVersion, "Python")
		// 	if err != nil {
		// 		return err
		// 	}

		// 	selectedRuntime = ret
		// } else if selectedLanguage == config.JavaScript || selectedLanguage == config.TypeScript {
		// 	nodeVersion, err := detectNodeVersion()
		// 	if err != nil {
		// 		return err
		// 	}

		// 	ret, err := askRuntimeVersion(nodeVersion, "Node")
		// 	if err != nil {
		// 		return err
		// 	}

		// 	selectedRuntime = ret
		// }

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
		}

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Println("Creating .upify folder...")
		fmt.Println("Saving configuration to .upify/config.yml...")
		fmt.Println("Done!")
		return nil
	},
}

func validateEntrypoint(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}

	return nil
}

func determinePackageManager(language lang.Language) (lang.PackageManager, error) {
	switch language {
	case lang.Python:
		return lang.Pip, nil
	case lang.JavaScript, lang.TypeScript:
		if _, err := os.Stat("yarn.lock"); err == nil {
			return lang.Yarn, nil
		}
		if _, err := os.Stat("package-lock.json"); err == nil {
			return lang.Npm, nil
		}
		return lang.Npm, nil
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

func determineLanguage(entrypoint string) (lang.Language, error) {
	ext := filepath.Ext(entrypoint)
	switch ext {
	case ".py":
		return lang.Python, nil
	case ".js":
		return lang.JavaScript, nil
	case ".ts":
		return lang.TypeScript, nil
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}
}

func detectPythonVersion() (string, error) {
	cmd := exec.Command("python", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	version := strings.TrimSpace(strings.TrimPrefix(string(output), "Python"))
	return version, nil
}

func detectNodeVersion() (string, error) {
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	version := strings.TrimSpace(strings.TrimPrefix(string(output), "v"))
	return version, nil
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

func askLanguage() (lang.Language, error) {
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

	return lang.Language(language), nil
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

func askRuntimeVersion(detectedVersion string, runtime string) (string, error) {
	message := fmt.Sprintf("Detected %s version: %s. Press Enter to confirm or type a different version:", runtime, detectedVersion)
	versionPrompt := &survey.Input{
		Message: message,
		Default: detectedVersion,
	}

	var version string
	if err := survey.AskOne(versionPrompt, &version); err != nil {
		return "", err
	}

	return version, nil
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
