package command

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

func init() {
	rootCommand.AddCommand(statisticsReCmd)
}

var (
	action    string // 操作
	confirm   string // 确认
	grayColor = "\033[90m"
)

// 统计数据重跑
var statisticsReCmd = &cobra.Command{
	Use:   "statistics-re",
	Short: "run statistics-re",
	Long:  `run statistics-re`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("Failed to initialize config: %v", err)
		}

		// 初始化日志系统
		if err := logger.Init(); err != nil {
			log.Fatalf("Failed to initialize logger: %v", err)
		}

		// 初始化雪花ID
		utils.InitIdGenerator()
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Logger.Sync()

		// 功能列表
		fmt.Printf("%s--------------------------------%s\n", blueColor, resetColor)
		fmt.Printf("%s 【1001】指定销售订单 %s\n", blueColor, resetColor)
		fmt.Printf("%s 【1002】指定充值订单 %s\n", blueColor, resetColor)
		fmt.Printf("%s 【1003】指定公司全部订单 %s\n", blueColor, resetColor)
		fmt.Printf("%s 【9999】全部公司全部订单 %s\n", blueColor, resetColor)
		fmt.Printf("%s--------------------------------%s\n", blueColor, resetColor)

		// 读取用户输入
		fmt.Printf("%s输入序号执行对应操作: %s", yellowColor, resetColor)
		fmt.Scanln(&action)

		// 根据输入执行对应操作
		switch action {
		case "1001":
			// 执行单个销售订单的统计重跑逻辑
			var companyUUID, orderNo string

			fmt.Printf("%s请输入商户uuid:%s ", yellowColor, resetColor)
			fmt.Scanln(&companyUUID)
			companyUUID = strings.TrimSpace(companyUUID)

			fmt.Printf("%s请输入销售订单号:%s ", yellowColor, resetColor)
			fmt.Scanln(&orderNo)
			orderNo = strings.TrimSpace(orderNo)

			// 根据 APP 表实例化数据库连接
			var company model.Company
			database.GetDBManager(config.Database).GetDB(0).Where("uuid = ? AND delete_time = 0", companyUUID).Find(&company).Limit(1)
			if company.Uuid == 0 {
				fmt.Printf("%s商户不存在%s\n", redColor, resetColor)
				return
			}

			// 确认商户和订单信息是否正确
			fmt.Printf("%s你输入的商户是:%s %s(%d)  %s销售订单号是:%s %s\n", blueColor, resetColor, company.Name, company.Uuid, blueColor, resetColor, orderNo)
			fmt.Printf("%s请确认商户和订单信息是否正确？(y/n):%s ", yellowColor, resetColor)
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Printf("%s已取消执行%s\n", yellowColor, resetColor)
				return
			}

			// 确认是否执行
			confirm = ""
			fmt.Printf("%s是否确认执行？(y/n):%s ", yellowColor, resetColor)
			fmt.Scanln(&confirm)
			if confirm == "y" || confirm == "Y" {
				fmt.Printf("%s开始执行商户 %s(%d) 下销售订单 %s 的统计...%s\n", greenColor, company.Name, company.Uuid, orderNo, resetColor)
				start := time.Now()
				err := statisticsReSaleOrder(company, orderNo)
				fmt.Printf("%s统计耗时：%v%s\n", blueColor, time.Since(start), resetColor)
				if err != nil {
					fmt.Printf("%s统计失败: %s%s\n", redColor, err, resetColor)
				} else {
					fmt.Printf("%s统计成功%s\n", greenColor, resetColor)
				}
			} else {
				fmt.Printf("%s已取消执行%s\n", yellowColor, resetColor)
			}
		case "1002":
			// 执行单个充值订单的统计重跑逻辑
			var companyUUID, orderNo string

			fmt.Printf("%s请输入公司uuid:%s ", yellowColor, resetColor)
			fmt.Scanln(&companyUUID)
			companyUUID = strings.TrimSpace(companyUUID)

			fmt.Printf("%s请输入充值订单号:%s ", yellowColor, resetColor)
			fmt.Scanln(&orderNo)
			orderNo = strings.TrimSpace(orderNo)

			// 根据 APP 表实例化数据库连接
			var company model.Company
			database.GetDBManager(config.Database).GetDB(0).Where("uuid = ? AND delete_time = 0", companyUUID).Find(&company).Limit(1)
			if company.Uuid == 0 {
				fmt.Printf("%s商户不存在%s\n", redColor, resetColor)
				return
			}

			// 确认商户和订单信息是否正确
			fmt.Printf("%s你输入的商户是:%s %s(%d)  %s充值订单号是:%s %s\n", blueColor, resetColor, company.Name, company.Uuid, blueColor, resetColor, orderNo)
			fmt.Printf("%s请确认商户和订单信息是否正确？(y/n):%s ", yellowColor, resetColor)
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Printf("%s已取消执行%s\n", yellowColor, resetColor)
				return
			}

			// 确认是否执行
			confirm = ""
			fmt.Printf("%s是否确认执行？(y/n):%s ", yellowColor, resetColor)
			fmt.Scanln(&confirm)
			if confirm == "y" || confirm == "Y" {
				fmt.Printf("%s开始执行商户 %s(%d) 下充值订单 %s 的统计...%s\n", greenColor, company.Name, company.Uuid, orderNo, resetColor)
				start := time.Now()
				err := statisticsReMemberRechargeOrder(company, orderNo)
				fmt.Printf("%s统计耗时：%v%s\n", blueColor, time.Since(start), resetColor)
				if err != nil {
					fmt.Printf("%s统计失败: %s%s\n", redColor, err, resetColor)
				} else {
					fmt.Printf("%s统计成功%s\n", greenColor, resetColor)
				}
			} else {
				fmt.Printf("%s已取消执行%s\n", yellowColor, resetColor)
			}
		case "1003":
			// 执行单个公司所有订单的统计重跑逻辑
			var companyUUID string
			fmt.Printf("%s请输入公司uuid:%s ", yellowColor, resetColor)
			fmt.Scanln(&companyUUID)
			companyUUID = strings.TrimSpace(companyUUID)

			// 根据 APP 表实例化数据库连接
			var companies []model.Company
			database.GetDBManager(config.Database).GetDB(0).Where("uuid = ? AND delete_time = 0", companyUUID).Find(&companies)
			if len(companies) == 0 {
				fmt.Printf("%s商户不存在%s\n", redColor, resetColor)
				return
			}

			// 如果商户存在多个，则提示用户检查脚本和数据是否正确
			if len(companies) > 1 {
				fmt.Printf("%s商户存在多个,请检查脚本和数据库是否正确%s\n", redColor, resetColor)
				return
			}

			// 确认商户和订单信息是否正确
			fmt.Printf("%s你输入的商户是:%s %s(%d)  %s\n", blueColor, resetColor, companies[0].Name, companies[0].Uuid, blueColor)
			confirm = ""
			fmt.Printf("%s请确认商户和订单信息是否正确？(y/n):%s ", yellowColor, resetColor)
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Printf("%s已取消执行%s\n", yellowColor, resetColor)
				return
			}

			// 确认是否执行
			confirm = ""
			fmt.Printf("%s是否确认执行？(y/n):%s ", yellowColor, resetColor)
			fmt.Scanln(&confirm)
			if confirm == "y" || confirm == "Y" {
				fmt.Printf("%s开始执行指定商户下所有订单的统计...%s\n", greenColor, resetColor)
				start := time.Now()
				err := statisticsReAllOrder(companies)
				fmt.Printf("%s统计耗时：%v%s\n", blueColor, time.Since(start), resetColor)
				if err != nil {
					fmt.Printf("%s统计失败: %s%s\n", redColor, err, resetColor)
				} else {
					fmt.Printf("%s统计成功%s\n", greenColor, resetColor)
				}
			} else {
				fmt.Printf("%s已取消执行%s\n", yellowColor, resetColor)
			}
		case "9999":
			var confirmation string
			// 根据 APP 表实例化数据库连接
			var companies []model.Company
			database.GetDBManager(config.Database).GetDB(0).Where("delete_time = 0").Find(&companies)
			if len(companies) == 0 {
				fmt.Printf("%s商户不存在,请检查脚本和数据库是否正确%s\n", redColor, resetColor)
				return
			}
			// 读取用户输入
			verificationCode := fmt.Sprintf("%05d", rand.Intn(90000)+10000) // 生成五位数字验证码
			fmt.Printf("%s请输入验证码 '%s' 继续，输入其他内容取消: %s", yellowColor, verificationCode, resetColor)
			fmt.Scanln(&confirmation)
			if confirmation != verificationCode {
				fmt.Printf("%s 验证码错误，已取消执行 %s\n", redColor, resetColor)
				return
			}
			// 执行全部公司的统计重跑逻辑
			fmt.Printf("%s是否确认执行？(y/n):%s ", yellowColor, resetColor)
			fmt.Scanln(&confirm)
			if confirm == "y" || confirm == "Y" {
				fmt.Printf("%s开始执行全部商户的统计...%s\n", greenColor, resetColor)
				start := time.Now()
				err := statisticsReAllOrder(companies)
				fmt.Printf("%s统计耗时：%v%s\n", blueColor, time.Since(start), resetColor)
				if err != nil {
					fmt.Printf("%s统计失败: %s%s\n", redColor, err, resetColor)
				} else {
					fmt.Printf("%s统计成功%s\n", greenColor, resetColor)
				}
			} else {
				fmt.Printf("%s已取消执行%s\n", yellowColor, resetColor)
			}
		default:
			// 输入无效，提示用户
			fmt.Printf("%s输入的序号无效，请重新输入%s\n", yellowColor, resetColor)
		}
	},
}

// 统计数据重跑 - 单个销售订单
func statisticsReSaleOrder(company model.Company, orderNo string) error {
	// 在上下文中添加公司uuid
	ctx := context.NewContext(context.WithCompanyUuid(company.Uuid))
	db := database.GetDBManager(config.Database).GetDB(company.Uuid)

	// 根据订单号查询订单
	var saleOrder model.SaleOrder
	db.Preload("SaleBill", func(db *gorm.DB) *gorm.DB {
		return db.Where("delete_time = 0")
	}).Where("order_no = ? AND delete_time = 0", orderNo).Find(&saleOrder).Limit(1)
	if saleOrder.Uuid == 0 || saleOrder.SaleBill == nil || saleOrder.SaleBill.Uuid == 0 {
		return errors.New("订单不存在")
	}

	// 重跑销售订单
	err := service.NewStatisticsSrv().SaveSale(ctx, service.SaveSaleReq{
		SaleBillUuid: saleOrder.SaleBill.Uuid,
		OnlyDelete:   false,
	})
	if err != nil {
		return err
	}

	return nil
}

// 统计数据重跑 - 单个充值订单
func statisticsReMemberRechargeOrder(company model.Company, orderNo string) error {
	// 在上下文中添加公司uuid
	ctx := context.NewContext(context.WithCompanyUuid(company.Uuid))
	db := database.GetDBManager(config.Database).GetDB(company.Uuid)

	// 根据订单号查询订单
	var memberRechargeOrder model.MemberRechargeOrder
	db.Where("order_no = ? AND delete_time = 0", orderNo).Find(&memberRechargeOrder).Limit(1)
	if memberRechargeOrder.Uuid == 0 {
		return errors.New("订单不存在")
	}

	// 重跑充值订单
	err := service.NewStatisticsSrv().SaveMember(ctx, service.SaveMemberReq{
		MemberRechargeOrderUuid: memberRechargeOrder.Uuid,
		OnlyDelete:              false,
	})
	if err != nil {
		return err
	}

	return nil
}

// 统计数据重跑 - 指定/全部公司全部订单
func statisticsReAllOrder(companies []model.Company) error {
	maxGoroutines := 50 // 最大并发数
	guard := make(chan struct{}, maxGoroutines)

	// 遍历公司，重跑所有订单
	for _, company := range companies {
		fmt.Printf("%s==============================%s\n", blueColor, resetColor)
		fmt.Printf("%s开始重新统计 - 商户 %s%s(%d)%s\n", blueColor, greenColor, company.Name, company.Uuid, resetColor)

		// 在上下文中添加公司uuid
		ctx := context.NewContext(context.WithCompanyUuid(company.Uuid))
		// 获取公司数据库连接
		db := database.GetDBManager(config.Database).GetDB(company.Uuid)

		// 先查总数，计算总页数
		pageSize := 10
		saleBillRepo := repository.NewSaleBillRepo(db)
		total, _ := saleBillRepo.GetCompleteTotal()
		fmt.Printf("%s 销售订单总数 - %d %s\n", blueColor, total, resetColor)
		pageCount := int((total + int64(pageSize) - 1) / int64(pageSize))
		ticker := time.NewTicker(time.Second)
		totalSuccess := 0
		totalFail := 0
		for pageNo := 1; pageNo <= pageCount; pageNo++ {
			<-ticker.C // 每秒处理一组
			saleBills, _, err := saleBillRepo.GetSaleBillListPage(pageNo, pageSize,
				repository.CommonRepo.WhereBySoftDelete(),
				repository.CommonRepo.WhereByStatus(constant.SaleBillStatusComplete),
			)
			if err != nil {
				fmt.Printf("%s获取销售订单分页失败: %v%s\n", redColor, err, resetColor)
				continue
			}

			var wg sync.WaitGroup
			pageSuccess := 0
			pageFail := 0
			for _, bill := range saleBills {
				wg.Add(1)
				guard <- struct{}{}
				utils.Go(func() {
					func(bill *model.SaleBill) {
						defer wg.Done()
						defer func() { <-guard }()
						err := service.NewStatisticsSrv().SaveSale(ctx, service.SaveSaleReq{
							SaleBillUuid: bill.Uuid,
							OnlyDelete:   false,
						})
						if err != nil {
							fmt.Printf("%s销售订单 %d 统计失败: %v%s\n", redColor, bill.Uuid, err, resetColor)
							pageFail++
						} else {
							pageSuccess++
						}
					}(bill)
				})

			}
			wg.Wait()
			totalSuccess += pageSuccess
			totalFail += pageFail

			startIndex := (pageNo-1)*pageSize + 1
			endIndex := startIndex + len(saleBills) - 1
			progress := float64(endIndex) / float64(total)
			barLen := 20
			filled := min(int(progress*float64(barLen)), barLen)
			// 进度条颜色优化
			barGreen := greenColor + strings.Repeat("#", filled) + resetColor
			barGray := grayColor + strings.Repeat("-", barLen-filled) + resetColor
			bar := barGreen + barGray
			percent := int(progress * 100)
			fmt.Printf("%s [%s] %d%% (%d/%d)%s\n", blueColor, bar, percent, endIndex, total, resetColor)
		}
		ticker.Stop()
		fmt.Printf("%s销售订单统计总成功: %d，总失败: %d%s\n\n", greenColor, totalSuccess, totalFail, resetColor)

		// 分页获取并处理所有的充值订单，避免一次性加载全部数据
		pageSize2 := 10
		memberRechargeOrderRepo := repository.NewMemberRechargeOrderRepo(db)
		total2, _ := memberRechargeOrderRepo.GetOrderCount(
			repository.CommonRepo.WhereBySoftDelete(),
			repository.CommonRepo.WhereByStatus(constant.RechargeOrderStatusPaid),
			repository.CommonRepo.WhereGtUuid(1000000),
		)
		fmt.Printf("%s 充值订单总数 - %d %s\n", blueColor, total2, resetColor)
		pageCount2 := int((total2 + int64(pageSize2) - 1) / int64(pageSize2))
		ticker2 := time.NewTicker(time.Second)
		totalSuccess2 := 0
		totalFail2 := 0
		for pageNo2 := 1; pageNo2 <= pageCount2; pageNo2++ {
			<-ticker2.C // 每秒处理一组
			rechargeOrders, _, err := memberRechargeOrderRepo.PaginateGetRechargeOrder(pageNo2, pageSize2,
				repository.CommonRepo.WhereBySoftDelete(),
				repository.CommonRepo.WhereByStatus(constant.RechargeOrderStatusPaid),
				repository.CommonRepo.WhereGtUuid(1000000),
			)
			if err != nil {
				fmt.Printf("%s获取充值订单分页失败: %v%s\n", redColor, err, resetColor)
				continue
			}

			var wg2 sync.WaitGroup
			pageSuccess2 := 0
			pageFail2 := 0
			for _, order := range rechargeOrders {
				wg2.Add(1)
				guard <- struct{}{}
				utils.Go(func() {
					func(order model.MemberRechargeOrder) {
						defer wg2.Done()
						defer func() { <-guard }()
						err := service.NewStatisticsSrv().SaveMember(ctx, service.SaveMemberReq{
							MemberRechargeOrderUuid: order.Uuid,
							OnlyDelete:              false,
						})
						if err != nil {
							fmt.Printf("%s充值订单 %d 统计失败: %v%s\n", redColor, order.Uuid, err, resetColor)
							pageFail2++
						} else {
							pageSuccess2++
						}
					}(order)
				})
			}
			wg2.Wait()
			totalSuccess2 += pageSuccess2
			totalFail2 += pageFail2

			startIndex2 := (pageNo2-1)*pageSize2 + 1
			endIndex2 := startIndex2 + len(rechargeOrders) - 1
			progress2 := float64(endIndex2) / float64(total2)
			barLen2 := 20
			filled2 := min(int(progress2*float64(barLen2)), barLen2)
			// 进度条颜色优化
			barGreen2 := greenColor + strings.Repeat("#", filled2) + resetColor
			barGray2 := grayColor + strings.Repeat("-", barLen2-filled2) + resetColor
			bar2 := barGreen2 + barGray2
			percent2 := int(progress2 * 100)
			fmt.Printf("%s [%s] %d%% (%d/%d)%s\n", blueColor, bar2, percent2, endIndex2, total2, resetColor)
		}
		ticker2.Stop()
		fmt.Printf("%s充值订单统计总成功: %d，总失败: %d%s\n", greenColor, totalSuccess2, totalFail2, resetColor)
		fmt.Printf("%s==============================%s\n\n", blueColor, resetColor)
	}

	return nil
}
