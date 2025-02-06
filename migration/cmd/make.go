package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"time"
	"ttpos-server-go/migration/util"
)

var (
	name     string
	filename string
)

func init() {
	rootCmd.AddCommand(makeCmd)
	makeCmd.Flags().StringVar(&name, "name", "", "migration file name")
}

var makeCmd = &cobra.Command{
	Use:   "make",
	Short: "Generate migration files",
	Long:  `Generate migration files`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if name == "" {
			util.PrintError("Name is required!")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		now := time.Now()
		timestamp := now.Format("20060102150405")
		filename = fmt.Sprintf("%s_%s.up.sql", timestamp, util.SnakeString(name))
		err := os.WriteFile(migrations+"/"+filename, []byte{}, os.ModePerm)
		if err != nil {
			util.PrintError("WriteFile error: " + err.Error() + "!")
			os.Exit(1)
		}
		filename = fmt.Sprintf("%s_%s.down.sql", timestamp, util.SnakeString(name))
		err = os.WriteFile(migrations+"/"+filename, []byte{}, os.ModePerm)
		if err != nil {
			util.PrintError("WriteFile error: " + err.Error() + "!")
			os.Exit(1)
		}
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		util.PrintInfo("Created migration [" + filename + "]")
	},
}
