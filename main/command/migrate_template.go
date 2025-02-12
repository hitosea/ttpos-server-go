package command

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

func init() {
	rootCommand.AddCommand(migrateTemplateCmd)
	migrateTemplateCmd.Flags().StringVar(&name, "name", "", "migration file name")
}

var (
	name       string
	filename   string
	migrations = "./migration/shop"
)

var migrateTemplateCmd = &cobra.Command{
	Use:   "migrate-template",
	Short: "Generate migration files",
	Long:  `Generate migration files`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if name == "" {
			PrintError("Name is required!")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		now := time.Now()
		timestamp := now.Format("20060102150405")
		filename = fmt.Sprintf("%s_%s.up.sql", timestamp, SnakeString(name))
		err := os.WriteFile(migrations+"/"+filename, []byte{}, os.ModePerm)
		if err != nil {
			PrintError("WriteFile error: " + err.Error() + "!")
			os.Exit(1)
		}
		filename = fmt.Sprintf("%s_%s.down.sql", timestamp, SnakeString(name))
		err = os.WriteFile(migrations+"/"+filename, []byte{}, os.ModePerm)
		if err != nil {
			PrintError("WriteFile error: " + err.Error() + "!")
			os.Exit(1)
		}
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		PrintInfo("Created migration [" + filename + "]")
	},
}

// SnakeString 驼峰转蛇形 XxYy to xx_yy , XxYY to xx_y_y
func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		// or通过ASCII码进行大小写的转化
		// 65-90（A-Z），97-122（a-z）
		//判断如果字母为大写的A-Z就在前面拼接一个_
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	//ToLower把大写字母统一转小写
	return strings.ToLower(string(data[:]))
}

// CamelString 蛇形转驼峰 xx_yy to XxYx  xx_y_y to XxYY
func CamelString(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

func PrintInfo(content string) {
	fmt.Println(GetInfoMsg(content))
}

func PrintError(msg string) {
	fmt.Printf("  \033[41;37m ERROR \033[0m " + msg + "\n")
}

func GetInfoMsg(content string) string {
	return fmt.Sprintf("  \033[44;37m INFO \033[0m " + content)
}

func GetSuccessMsg(msg string) string {
	return fmt.Sprintf(" \033[1;32m" + msg + "\033[0m")
}

func GetSuccessMsgWithPrefix(prefix, msg string) string {
	return fmt.Sprintf(prefix + "\033[1;32m" + msg + "\033[0m")
}

func GetWarningMsg(msg string) string {
	return fmt.Sprintf(" \033[1;33m" + msg + "\033[0m")
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}
	return false
}
