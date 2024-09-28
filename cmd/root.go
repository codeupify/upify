package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "upify",
	Short: "Upify helps you deploy applications to the cloud effortlessly.",
	Long:  `Upify is a CLI tool designed to simplify cloud deployments for various runtimes and platforms.`,
}

func Execute() error {
	return rootCmd.Execute()
}
