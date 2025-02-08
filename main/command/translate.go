package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

type Message struct {
	ID    string
	Value string
}

// YoudaoResponse 使用有道翻译API翻译文本
type YoudaoResponse struct {
	ErrorCode   string   `json:"errorCode"`
	Query       string   `json:"query"`
	Translation []string `json:"translation"`
}

func init() {
	rootCommand.AddCommand(translateCmd)
}

var translateCmd = &cobra.Command{
	Use:   "translate",
	Short: "开始翻译",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("-------translate start-------\n")
	},
}
