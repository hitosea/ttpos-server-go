package command

import (
	"fmt"
	"ttpos-server-go/pkg/utils"

	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(printCmd)
}

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Print the printer type",
	Run: func(cmd *cobra.Command, args []string) {
		printerType := PrinterType{
			PrinterType: "test",
		}
		fmt.Println(utils.ToJsonString(printerType))
	},
}

type PrinterType struct {
	PrinterType string `json:"printer_type"`
}
