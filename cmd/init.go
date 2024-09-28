package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var (
	runtime    string
	platform   string
	framework  string
	entrypoint string
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&runtime, "runtime", "r", "", "Runtime of your project (node, python)")
	initCmd.Flags().StringVarP(&framework, "framework", "f", "", "Framework used (none, flask, express)")
	initCmd.Flags().StringVarP(&platform, "platform", "p", "", "Platform to deploy to (aws-lambda, gcp-cloudrun)")
	initCmd.Flags().StringVarP(&entrypoint, "entrypoint", "e", "", "Path to your main application file (e.g., app.js, main.py)")
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize your project for deployment",
	Run: func(cmd *cobra.Command, args []string) {
		manualMode := cmd.Flags().NFlag() > 0

		if manualMode {
			if runtime == "" || framework == "" || platform == "" || entrypoint == "" {
				fmt.Println("Error: When providing manual parameters, all of the following are required:")
				fmt.Println("  -r, --runtime")
				fmt.Println("  -f, --framework")
				fmt.Println("  -p, --platform")
				fmt.Println("  -e, --entrypoint")
				return
			}

			if (runtime == "python" && framework != "none" && framework != "flask") ||
				(runtime == "node" && framework != "none" && framework != "express") {
				fmt.Printf("Error: Invalid runtime and framework combination: %s and %s\n", runtime, framework)
				return
			}
		} else {
			questions := []*survey.Question{
				{
					Name:     "runtime",
					Prompt:   &survey.Select{Message: "Choose a runtime:", Options: []string{"node", "python"}},
					Validate: survey.Required,
				},
			}

			answers := struct {
				Runtime    string
				Framework  string
				Platform   string
				Entrypoint string
			}{}

			err := survey.Ask(questions[:1], &answers)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			runtime = answers.Runtime

			var frameworkOptions []string
			if runtime == "node" {
				frameworkOptions = []string{"none", "express"}
			} else if runtime == "python" {
				frameworkOptions = []string{"none", "flask"}
			}

			questions = append(questions,
				&survey.Question{
					Name:     "framework",
					Prompt:   &survey.Select{Message: "Choose a framework (or 'none' for standalone scripts/apps):", Options: frameworkOptions},
					Validate: survey.Required,
				},
				&survey.Question{
					Name:     "platform",
					Prompt:   &survey.Select{Message: "Choose a platform:", Options: []string{"aws-lambda", "gcp-cloudrun"}},
					Validate: survey.Required,
				},
				&survey.Question{
					Name:   "entrypoint",
					Prompt: &survey.Input{Message: "Enter the path to your main application file (e.g., app.js, main.py):"},
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
			)

			err = survey.Ask(questions[1:], &answers)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			framework = answers.Framework
			platform = answers.Platform
			entrypoint = answers.Entrypoint
		}

		fmt.Printf("Initializing project with runtime: %s, framework: %s, platform: %s, entrypoint: %s\n",
			runtime, framework, platform, entrypoint)
	},
}
