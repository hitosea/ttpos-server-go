package command

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	orderNo      string
	statusBefore string
	statusAfter  string
)

// skootar订单状态映射
var skootarStatusMap = map[string]string{
	"0":  "Canceled - 已取消",
	"1":  "New - 新订单",
	"2":  "Under review - 正在审核中",
	"3":  "Mission created - 审核通过",
	"4":  "Matching - 正在匹配司机",
	"5":  "Assigned - 已分配司机",
	"6":  "OTW to pickup - 正在前往取货",
	"7":  "OTW to delivery - 正在前往送货",
	"8":  "Pending - 等待修改",
	"9":  "Completed - 已完成",
	"10": "Invoice Created - 已生成发票",
	"11": "Paid - 已支付服务费",
}

func init() {
	rootCommand.AddCommand(skootarStatusCmd)
	skootarStatusCmd.Flags().StringVar(&companyIdStr, "company-id", "", "公司UUID")
	skootarStatusCmd.Flags().StringVar(&orderNo, "order-no", "", "订单号")
}

// displayStatusOptions 显示状态选项
func displayStatusOptions() {
	fmt.Printf("%s ==================== Skootar 订单状态列表 ====================%s\n", blueColor, resetColor)
	for i := 0; i <= 11; i++ {
		status := fmt.Sprintf("%d", i)
		if desc, exists := skootarStatusMap[status]; exists {
			fmt.Printf("%s [%s] %s %s\n", yellowColor, status, desc, resetColor)
		}
	}
	fmt.Printf("%s =========================================================%s\n", blueColor, resetColor)
}

var skootarStatusCmd = &cobra.Command{
	Use:   "skootar-update-status",
	Short: "skootar-update-status",
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("Failed to initialize config: %v", err)
		}

		// 初始化日志系统
		if err := logger.Init(); err != nil {
			log.Fatalf("Failed to initialize logger: %v", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Logger.Sync()

		// 读取用户输入
		if companyIdStr == "" {
			fmt.Printf("%s 输入公司UUID (默认8609817471094784):  %s", blueColor, resetColor)
			fmt.Scanln(&companyIdStr)
			if companyIdStr == "" {
				companyIdStr = "8609817471094784"
			}
		}

		// 检查公司ID是否为有效数字
		if _, err := fmt.Sscanf(companyIdStr, "%d", &companyUuid); err != nil {
			fmt.Printf("%s 错误: 公司ID必须是有效的数字，当前值: %s%s\n", redColor, companyIdStr, resetColor)
			return
		}

		// 根据 APP 表实例化数据库连接
		var companies []model.Company
		if err := database.GetDBManager(config.Database).GetDB(0).Scopes(repository.NotDeleted).Debug().Find(&companies).Error; err != nil {
			log.Fatalf("Error querying companies: %s", err)
		}

		// 读取用户输入
		if orderNo == "" {
			fmt.Printf("%s 输入订单号进行更新状态: %s", blueColor, resetColor)
			fmt.Scanln(&orderNo)
			if orderNo == "" {
				fmt.Printf("%s 订单号不能为空 %s\n", redColor, resetColor)
				return
			}
		}

		// 获取对应的订单
		fmt.Println("skootar-update-status")

		// 执行查询和请求发送
		if err := processOrderStatusUpdate(); err != nil {
			fmt.Printf("%s 处理失败: %v %s\n", redColor, err, resetColor)
			return
		}

		fmt.Printf("%s 订单状态更新请求发送成功！ %s\n", greenColor, resetColor)
	},
}

// processOrderStatusUpdate 处理订单状态更新
func processOrderStatusUpdate() error {
	// 获取数据库连接
	dbManager := database.GetDBManager(config.Database)
	db := dbManager.GetDB(companyUuid)

	// 创建会员订单仓库
	memberSaleOrderRepo := repository.NewMemberSaleOrderRepo(db)

	// 根据订单号查询会员订单
	memberSaleOrder, err := memberSaleOrderRepo.GetMemberSaleOrder(
		repository.CommonRepo.WhereBySoftDelete(),
		repository.CommonRepo.WhereByOrderNo(orderNo),
	)
	if err != nil {
		return fmt.Errorf("查询订单失败: %v", err)
	}

	// 检查是否找到订单
	if memberSaleOrder.Uuid == 0 {
		return fmt.Errorf("未找到订单号为 %s 的订单", orderNo)
	}

	// 检查是否有关联订单号
	if memberSaleOrder.RelatedOrderNo == "" {
		return fmt.Errorf("订单 %s 没有关联的外送订单号", orderNo)
	}

	fmt.Printf("%s 找到订单: %s, 关联外送订单号: %s %s\n", greenColor, orderNo, memberSaleOrder.RelatedOrderNo, resetColor)

	// 显示状态选项
	displayStatusOptions()

	// 让用户输入变更前状态
	fmt.Printf("%s 输入变更后状态:  %s", blueColor, resetColor)
	fmt.Scanln(&statusAfter)
	if statusAfter == "" {
		return fmt.Errorf("变更后状态不能为空")
	}

	statusAfterInt, _ := strconv.Atoi(statusAfter)
	statusBeforeInt := statusAfterInt - 1
	statusBefore = strconv.Itoa(statusBeforeInt)

	// 验证状态合法性
	if _, exists := skootarStatusMap[statusAfter]; !exists {
		return fmt.Errorf("变更后状态 %s 不合法，请输入0-11之间的数字", statusAfter)
	}

	// 显示状态变更信息
	fmt.Printf("%s 状态变更: [%s] %s -> [%s] %s %s\n",
		greenColor, statusBefore, skootarStatusMap[statusBefore],
		statusAfter, skootarStatusMap[statusAfter], resetColor)

	// 构造请求数据
	requestData := map[string]interface{}{
		"callback_desc":   "status_change",
		"status_datetime": time.Now().Format("2006-01-02 15:04:05"),
		"job_id":          memberSaleOrder.RelatedOrderNo,
		"status_before":   statusBeforeInt,
		"status_after":    statusAfterInt,
	}

	// 发送HTTP请求
	// 从环境变量获取 takeout 服务 URL
	takeoutServiceURL := viper.GetString("TAKEOUT_SERVICE_URL")
	if takeoutServiceURL == "" {
		takeoutServiceURL = "http://127.0.0.1:14031" // 默认值
		fmt.Printf("%s 未配置TAKEOUT_SERVICE_URL环境变量，使用默认值: %s %s\n", yellowColor, takeoutServiceURL, resetColor)
	} else {
		fmt.Printf("%s 使用TAKEOUT_SERVICE_URL: %s %s\n", greenColor, takeoutServiceURL, resetColor)
	}
	// 确保URL不以斜杠结尾
	takeoutServiceURL = strings.TrimSuffix(takeoutServiceURL, "/")

	url := takeoutServiceURL + "/callback/skootar/status"
	headers := map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}

	fmt.Printf("%s 正在发送状态更新请求... %s\n", blueColor, resetColor)
	fmt.Printf("请求URL: %s\n", url)
	fmt.Printf("请求数据: %+v\n", requestData)

	response, err := utils.HttpPost(url, requestData, headers)
	if err != nil {
		return fmt.Errorf("发送HTTP请求失败: %v", err)
	}

	fmt.Printf("%s 请求响应: %+v %s\n", greenColor, response, resetColor)

	return nil
}
