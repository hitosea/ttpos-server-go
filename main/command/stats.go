package command

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"ttpos-server-go/config"

	"github.com/spf13/cobra"
)

// MemoryStatsResponse 内存统计响应结构
type MemoryStatsResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// displayMemoryStats 显示内存统计信息
func displayMemoryStats(body []byte) {
	var memResponse MemoryStatsResponse
	if err := json.Unmarshal(body, &memResponse); err != nil {
		fmt.Printf("解析内存统计信息失败: %s\n", err.Error())
		return
	}

	data := memResponse.Data
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("          内存统计信息")
	fmt.Printf("更新时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 50))

	// 基本统计信息
	fmt.Println("\n📊 基本统计:")
	fmt.Printf("  协程数量: %v\n", data["goroutines"])
	fmt.Printf("  垃圾回收次数: %v\n", data["num_gc"])
	fmt.Printf("  分配次数: %v\n", data["mallocs"])
	fmt.Printf("  释放次数: %v\n", data["frees"])

	// 内存使用情况
	fmt.Println("\n💾 内存使用情况:")
	fmt.Printf("  系统内存分配: %v KB\n", data["alloc_kb"])
	fmt.Printf("  总分配内存: %v KB\n", data["total_alloc_kb"])
	fmt.Printf("  系统内存: %v KB\n", data["sys_kb"])

	// 堆内存详情
	fmt.Println("\n🗃️  堆内存详情:")
	fmt.Printf("  堆内存大小: %v KB\n", data["heap_alloc_kb"])
	fmt.Printf("  堆内存系统: %v KB\n", data["heap_sys_kb"])
	fmt.Printf("  堆内存空闲: %v KB\n", data["heap_idle_kb"])
	fmt.Printf("  堆内存使用: %v KB\n", data["heap_inuse_kb"])
	fmt.Printf("  堆内存对象: %v\n", data["heap_objects"])

	// 其他内存
	fmt.Println("\n📚 其他内存:")
	fmt.Printf("  栈内存: %v KB\n", data["stack_inuse_kb"])
	fmt.Printf("  MCache使用: %v KB\n", data["mcache_inuse_kb"])
	fmt.Printf("  MSpan使用: %v KB\n", data["mspan_inuse_kb"])

	// GC性能
	fmt.Println("\n⚡ GC性能:")
	fmt.Printf("  GC CPU使用率: %.6f\n", data["gc_cpu_fraction"])
	fmt.Printf("  总暂停时间: %v ns\n", data["pause_total_ns"])
	fmt.Printf("  下次GC目标: %v KB\n", data["next_gc_kb"])

	// ImgFont状态
	fmt.Println("\n🖼️ ImgFont状态:")
	if imgFontData, ok := data["imgFontSemaphore"].(map[string]interface{}); ok {
		fmt.Printf("  最大并发: %v\n", imgFontData["max_concurrent"])
		fmt.Printf("  活跃数量: %v\n", imgFontData["active_count"])
		fmt.Printf("  可用数量: %v\n", imgFontData["available"])
		fmt.Printf("  信号量长度: %v\n", imgFontData["semaphore_length"])
		fmt.Printf("  字体库数量: %v\n", imgFontData["font_count"])
		fmt.Printf("  字体宽度数量: %v\n", imgFontData["width_cache_count"])
		fmt.Printf("  访问缓存计数: %v\n", imgFontData["access_cache_count"])
		if memoryUsage, ok := imgFontData["memory_usage"].(float64); ok {
			fmt.Printf("  字体管理器内存使用: %.2f M\n", memoryUsage/1024/1024)
		} else if memoryUsage, ok := imgFontData["memory_usage"].(int64); ok {
			fmt.Printf("  字体管理器内存使用: %.2f M\n", float64(memoryUsage)/1024/1024)
		} else {
			fmt.Printf("  字体管理器内存使用: %v\n", imgFontData["memory_usage"])
		}
	} else {
		fmt.Println("  ImgFont状态数据不可用")
	}

	fmt.Println(strings.Repeat("=", 50))
}

func init() {
	rootCommand.AddCommand(statsCmd)
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "实时监控系统内存和ImgFont状态",
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("Failed to initialize config: %v", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// 在这里调用接口获取内存统计信息
		req, err := http.NewRequest("GET", "http://127.0.0.1:8080/api/v1/internal/memory/stats", nil)
		if err != nil {
			fmt.Println("创建HTTP请求失败: " + err.Error())
			return
		}

		// 添加API密钥头部
		req.Header.Set("X-API-KEY", config.JWT.Secret)

		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			fmt.Println("获取内存统计信息失败: " + err.Error())
			return
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("读取内存统计信息失败: " + err.Error())
			return
		}

		if response.StatusCode != 200 {
			fmt.Printf("获取内存统计信息失败，状态码: %d, 响应: %s\n", response.StatusCode, string(body))
			return
		}

		// 开始动态更新内存统计信息
		fmt.Println("\n开始动态监控内存统计信息... (按 Ctrl+C 退出)")
		fmt.Println("==================================================")

		// 创建定时器，每秒更新一次
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		// 创建信号通道
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// 立即显示一次
		displayMemoryStats(body)

		// 循环更新
		for {
			select {
			case <-sigChan:
				fmt.Println("\n\n正在退出监控...")
				return
			case <-ticker.C:
				// 获取最新的内存统计信息
				req, err := http.NewRequest("GET", "http://127.0.0.1:8080/api/v1/internal/memory/stats", nil)
				if err != nil {
					fmt.Printf("\n创建HTTP请求失败: %s\n", err.Error())
					continue
				}

				req.Header.Set("X-API-KEY", config.JWT.Secret)

				client := &http.Client{}
				response, err := client.Do(req)
				if err != nil {
					fmt.Printf("\n获取内存统计信息失败: %s\n", err.Error())
					continue
				}
				defer response.Body.Close()

				body, err := io.ReadAll(response.Body)
				if err != nil {
					fmt.Printf("\n读取内存统计信息失败: %s\n", err.Error())
					continue
				}

				if response.StatusCode != 200 {
					fmt.Printf("\n获取内存统计信息失败，状态码: %d\n", response.StatusCode)
					continue
				}

				// 清屏并显示新数据
				fmt.Print("\033[2J\033[H") // 清屏
				displayMemoryStats(body)
			}
		}
	},
}
