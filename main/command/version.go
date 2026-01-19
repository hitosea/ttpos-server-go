package command

import (
	"fmt"
	"os"
	"strings"
	"time"
	"ttpos-server-go/config"

	"github.com/spf13/cobra"
)

var (
	newVersion   string
	newCommit    string
	newBuildTime string
)

func init() {
	versionCmd.Flags().StringVar(&newVersion, "version", "", "设置新的版本号")
	versionCmd.Flags().StringVar(&newCommit, "commit", "", "设置新的提交SHA")
	versionCmd.Flags().StringVar(&newBuildTime, "build-time", "", "设置新的构建时间")
	rootCommand.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		// 如果提供了任何参数，则更新版本信息
		if newVersion != "" || newCommit != "" || newBuildTime != "" {
			updateVersion()
			fmt.Println("版本已更新为: " + config.Version + "/" + config.CommitSHA + ", build at " + config.BuildTime)
		} else {
			fmt.Println("ttpos-server-go version: " + config.Version + "/" + config.CommitSHA + ", build at " + config.BuildTime)
		}
	},
}

// 更新版本信息
func updateVersion() {
	// 版本文件路径
	versionFilePath := "config/version.go"

	// 读取当前版本文件内容
	content, err := os.ReadFile(versionFilePath)
	if err != nil {
		fmt.Printf("读取版本文件失败: %v\n", err)
		return
	}

	fileContent := string(content)

	// 更新版本号
	if newVersion != "" {
		fileContent = updateVersionField(fileContent, "Version", newVersion)
	}

	// 更新提交SHA
	if newCommit != "" {
		fileContent = updateVersionField(fileContent, "CommitSHA", newCommit)
	}

	// 更新构建时间
	if newBuildTime != "" {
		fileContent = updateVersionField(fileContent, "BuildTime", newBuildTime)
	} else if newVersion != "" || newCommit != "" {
		// 如果没有指定构建时间但更新了其他字段，自动更新构建时间为当前时间
		currentTime := time.Now().Format("2006-01-02")
		fileContent = updateVersionField(fileContent, "BuildTime", currentTime)
	}

	// 写入更新后的内容
	err = os.WriteFile(versionFilePath, []byte(fileContent), 0644)
	if err != nil {
		fmt.Printf("更新版本文件失败: %v\n", err)
		return
	}

	// 更新内存中的版本信息
	if newVersion != "" {
		config.Version = newVersion
	}
	if newCommit != "" {
		config.CommitSHA = newCommit
	}
	if newBuildTime != "" {
		config.BuildTime = newBuildTime
	} else if newVersion != "" || newCommit != "" {
		config.BuildTime = time.Now().Format("2006-01-02")
	}
}

// 更新版本文件中的特定字段
func updateVersionField(content, field, value string) string {
	// 查找字段行
	fieldPattern := field + "\\s*=\\s*\"[^\"]*\""
	replacement := field + " = \"" + value + "\""

	// 替换字段值
	return strings.Replace(content, findFieldLine(content, fieldPattern), replacement, 1)
}

// 查找匹配的字段行
func findFieldLine(content, pattern string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, pattern[:strings.Index(pattern, "\\")]) {
			return strings.TrimSpace(line)
		}
	}
	return ""
}
