package cmd

import (
	"github.com/spf13/cobra"
)

var version = "0.96.0"

var rootCmd = &cobra.Command{
	Use:     "upify",
	Short:   "Upify helps you quickly and easily deploy apps in the cloud",
	Long:    `Upify is a platform and cloud agnostic CLI tool designed to simplify cloud deployments`,
	Version: version,
}

func Execute() error {
	rootCmd.SetVersionTemplate("Upify version: {{.Version}}\n") // Customize the version output if needed
	return rootCmd.Execute()
}
