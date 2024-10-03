package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "upify",
	Short: "Upify helps you quickly and easily deploy apps in the cloud",
	Long:  `Upify is a platform and cloud agnostic CLI tool designed to simplify cloud deployments`,
}

func Execute() error {
	return rootCmd.Execute()
}
