package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var migrations = "./migration/files"

var rootCmd = &cobra.Command{
	Use:   "migrate",
	Short: "database migration",
	Long:  "Show how to use database migration",
	PreRun: func(cmd *cobra.Command, args []string) {

	},
	Run: func(cmd *cobra.Command, args []string) {
	},
	PostRun: func(cmd *cobra.Command, args []string) {
	},
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
