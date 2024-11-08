package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/framework"
	"github.com/codeupify/upify/internal/infra"
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

			ret, err := askEntrypoint("Enter the relative path to your main Flask application file, typically where Flask is instantiated (e.g., app.py or main.py):", "app.py")
			if err != nil {
				return err
			}
			entrypoint = ret

			ret, err = askAppVar("Enter the name of the Flask app variable (the instance of Flask() used to start your app):", "app")
			if err != nil {
				return err
			}
			appVar = ret

			selectedFramework = framework.Framework(frameworkStr)
			selectedLanguage = lang.Python

		case "express":

			ret, err := askEntrypoint("Enter the relative path to your main Express application file, typically where the Express is instantiated (e.g., app.js or index.js):", "app.js")
			if err != nil {
				return err
			}
			entrypoint = ret

			ret, err = askAppVar("Enter the name of the exported Express app instance variable:", "app")
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

		if err := infra.AddHandlerFile(cfg); err != nil {
			return err
		}

		if entrypoint == "" {
			if err := infra.AddMainFile(cfg); err != nil {
				return err
			}
		}

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		err = infra.AddEnvironmentFile()
		if err != nil {
			return err
		}

		fmt.Println("Done!")

		if entrypoint == "" {
			mainPath := infra.GetMainPath(selectedLanguage)
			fmt.Printf("\n\033[1mImportant!\033[0m To finish the configuration wrap your script by modifying the file at \033[1m%s\033[0m\n", mainPath)
		}

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
	name := filepath.Base(cwd)

	re := regexp.MustCompile(`[^a-z0-9\-]+`)
	sanitized := re.ReplaceAllString(strings.ToLower(name), "-")

	sanitized = strings.Trim(sanitized, "-")

	return sanitized, nil
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

func validateProjectName(name string) error {
	validName := regexp.MustCompile(`^[a-z0-9\-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("project name '%s' contains invalid characters", name)
	}

	if len(name) > 64 {
		return fmt.Errorf("project name '%s' is too long (maximum 64 characters)", name)
	}

	return nil
}

func askProjectName() (string, error) {
	for {
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

		if err := validateProjectName(projectName); err != nil {
			fmt.Printf("Invalid project name: %v\n", err)
			fmt.Println("Allowed characters: lowercase letters, numbers, and hyphens (-). Please try again.")
			continue
		}

		return projectName, nil
	}
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
