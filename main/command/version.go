package command

import (
	"fmt"
	"ttpos-server-go/version"

	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ttpos-server-go version: " + version.Version + "/" + version.CommitSHA + ", build at " + version.BuildTime)
	},
}
