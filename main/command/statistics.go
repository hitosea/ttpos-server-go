package command

import (
	"fmt"
	"log"
	"sync"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(statisticsCmd)
}

// 数据迁移
var statisticsCmd = &cobra.Command{
	Use:   "statistics",
	Short: "run statistics",
	Long:  `run statistics`,
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
		fmt.Printf("%s 输入要统计的公司UUID继续, 输入1001表示全部: %s", blueColor, resetColor)
		fmt.Scanln(&companyIdStr)
		if companyIdStr == "" {
			fmt.Printf("%s 统计已取消 %s\n", redColor, resetColor)
			return
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

		//
		isExist := false
		for _, company := range companies {
			if companyUuid != 1001 && companyUuid != company.Uuid {
				continue
			}
			isExist = true

			fmt.Printf("%s 开始重新统计 - 公司 %d %s\n", blueColor, company.Uuid, resetColor)

			// 在上下文中添加公司uuid
			ctx := context.NewContext(context.WithCompanyUuid(company.Uuid))
			db := database.GetDBManager(config.Database).GetDB(company.Uuid)

			// 先查总数，计算总页数
			pageSize := 1000
			pageNo := 1
			saleBillRepo := repository.NewSaleBillRepo(db)
			total, _ := saleBillRepo.GetCompleteTotal()
			// 输出总数
			fmt.Printf("%s 总数 - %d %s\n", blueColor, total, resetColor)
			//
			pageCount := int((total + int64(pageSize) - 1) / int64(pageSize))
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for pageNo = 1; pageNo <= pageCount; pageNo++ {
				<-ticker.C // 每秒处理一组
				saleBills, _, err := saleBillRepo.GetSaleBillListPage(pageNo, pageSize,
					repository.CommonRepo.WhereBySoftDelete(),
					repository.CommonRepo.WhereByStatus(constant.SaleBillStatusComplete),
				)
				if err != nil {
					continue
				}
				var wg sync.WaitGroup
				for _, bill := range saleBills {
					wg.Add(1)
					go func(bill *model.SaleBill) {
						defer wg.Done()
						service.NewStatisticsSrv().SaveSale(ctx, service.SaveSaleReq{
							SaleBillUuid: bill.Uuid,
							OnlyDelete:   false,
						})
					}(bill)
				}
				wg.Wait() // 当前页全部处理完再进入下一页
			}

			// 分页获取并处理所有的充值订单，避免一次性加载全部数据
			pageSize2 := 1000
			pageNo2 := 1
			memberRechargeOrderRepo := repository.NewMemberRechargeOrderRepo(db)
			// 获取总数，计算页数
			total2, _ := memberRechargeOrderRepo.GetOrderCount()
			pageCount2 := int((total2 + int64(pageSize2) - 1) / int64(pageSize2))
			ticker2 := time.NewTicker(time.Second)
			defer ticker2.Stop()
			for pageNo2 = 1; pageNo2 <= pageCount2; pageNo2++ {
				<-ticker2.C // 每秒处理一组
				rechargeOrders, _, err := memberRechargeOrderRepo.PaginateGetRechargeOrder(pageNo2, pageSize2)
				if err != nil {
					continue
				}
				var wg2 sync.WaitGroup
				for _, order := range rechargeOrders {
					wg2.Add(1)
					go func(order model.MemberRechargeOrder) {
						defer wg2.Done()
						service.NewStatisticsSrv().SaveMember(ctx, service.SaveMemberReq{
							MemberRechargeOrderUuid: order.Uuid,
							OnlyDelete:              false,
						})
					}(order)
				}
				wg2.Wait() // 当前页全部处理完再进入下一页
			}
		}

		// 检查公司ID是否存在
		if !isExist {
			fmt.Printf("%s 错误: 公司ID不存在，当前值: %s%s\n", redColor, companyIdStr, resetColor)
		} else {
			fmt.Printf("%s 统计完成 %s\n", greenColor, resetColor)
		}
	},
}
