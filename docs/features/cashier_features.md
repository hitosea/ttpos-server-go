# 需求设计文档：收银端核心功能

### 1. 简介

本文档旨在描述 TTPOS-Server-Go 项目中收银端（Cashier）模块的核心功能。收银端是餐饮系统中用于日常收银操作的关键部分，涵盖了订单管理、基础信息查询、系统设置、交班管理、打印功能、认证鉴权、自助餐管理、呼叫管理、桌台管理、H5 订单管理、即时订单（点餐）管理、会员外送订单管理、会员和充值订单管理、历史订单管理、打印记录管理以及产品管理等多个方面。

### 2. 功能模块划分

根据 `main/app/api/v1/cashier` 目录下的文件分析，收银端主要包含以下功能模块：

#### 2.1 基础信息与系统设置 (由 `cashier_base.go` 提供)

**概述**: 提供收银端运行所需的基础信息、系统配置管理以及相关的安全验证功能。

**详细功能点**:

*   **基础信息**:
    *   获取收银端基本信息 (如门店信息、员工信息等)。
    *   获取多语言配置。
    *   获取副屏广告信息。
*   **安全验证**:
    *   验证钱箱密码。
    *   验证高级密码 (用于敏感操作)。
    *   验证锁屏密码。
*   **系统设置**:
    *   检查收银端版本更新。
    *   获取支付方式列表 (可按类型筛选)。
    *   修改接单设置 (例如是否自动接单)。
    *   修改会员订单接单设置 (例如是否自动接会员外送订单)。
    *   修改系统通用设置。
    *   获取收银端所有设置信息。
*   **辅助功能**:
    *   获取退菜原因列表。
    *   获取免单/赠菜原因列表。
    *   获取待打印数据。
    *   获取打开钱箱的打印机配置。
    *   USB打印机上报功能 (上报打印状态或数据)。
    *   解密活动二维码 (用于营销活动)。

#### 2.2 交班管理 (由 `cashier_base.go` 提供)

**概述**: 支持收银员进行班次交接，包括交班信息的查询、提交、现金操作以及交班单据的打印。

**详细功能点**:

*   **交班信息**:
    *   获取当前班次交班信息 (如销售额、收款方式汇总、钱箱金额等)。
    *   提交交班。
    *   交班取钱操作。
    *   交班存钱操作。
    *   打印交班单据。
*   **报备管理**:
    *   获取报备信息 (如异常情况、库存盘点等)。
    *   提交报备信息。

#### 2.3 订单管理 (由 `cashier_order.go` 提供)

**概述**: 负责收银端订单的创建、查询、修改、取消、退款、反结账、删除以及打印功能。

**详细功能点**:

*   **订单查询**:
    *   获取订单列表 (支持分页、筛选)。
    *   获取订单详情。
*   **订单操作**:
    *   取消订单。
    *   获取退款信息 (用于退菜/退单前的预览)。
    *   退款订单 (支持整单退款或部分退款)。
    *   重新退款 (处理退款失败后的重试)。
    *   获取反结账弹窗信息 (反结账前的确认)。
    *   反结账。
    *   删除订单。
*   **订单状态判断**:
    *   判断订单是否可关闭 (例如是否有未完成的菜品)。
*   **打印功能**:
    *   打印小票。
    *   打印发票。
    *   获取发票信息。

#### 2.4 认证与鉴权 (由 `cashier_auth.go` 提供)

**概述**: 处理收银员的登录、刷新令牌和退出登录等认证相关功能。

**详细功能点**:

*   **登录**: 收银员通过账号密码登录系统，获取访问令牌 (Token) 和刷新令牌 (RefreshToken)。
*   **刷新令牌**: 使用刷新令牌获取新的访问令牌，保持登录状态。
*   **退出登录**: 终止当前会话，使访问令牌失效。

#### 2.5 自助餐管理 (由 `cashier_buffet.go` 提供)

**概述**: 提供收银端自助餐相关的查询功能。

**详细功能点**:

*   **自助餐列表**: 获取所有自助餐方案的列表。
*   **自助餐延迟列表**: 获取自助餐延迟的列表。

#### 2.6 呼叫管理 (由 `cashier_call.go` 提供)

**概述**: 处理收银端接收到的呼叫信息，包括异常打印和未处理呼叫，并支持相关操作。

**详细功能点**:

*   **异常打印**:
    *   获取异常打印列表。
    *   删除异常打印记录。
    *   重新打印。
*   **未处理呼叫**:
    *   获取未处理呼叫列表 (支持分页)。
    *   获取未处理呼叫的数量。
    *   获取未处理呼叫的通知信息。
    *   处理呼叫 (将未处理呼叫标记为已处理)。

#### 2.7 桌台管理 (由 `cashier_desk.go` 提供)

**概述**: 收银端桌台管理的核心模块，涵盖了桌台的生命周期管理、订单商品的精细化操作、结账流程、会员优惠以及与 ERP 系统和打印相关的集成。

**详细功能点**:

*   **桌台信息**:
    *   获取桌台区域和类型列表。
    *   获取桌台列表 (支持分页、筛选)。
    *   获取桌台详情。
*   **桌台操作**:
    *   开台 (创建桌台订单)。
    *   关闭桌台。
    *   清台 (完成桌台)。
    *   切换桌台 (转台)。
    *   合并桌台。
    *   取消桌台订单。
    *   订单解锁。
*   **购物车与订单商品操作**:
    *   查询桌台购物车信息 (包含 H5 订单商品)。
    *   向购物车添加商品。
    *   向购物车添加套餐。
    *   查询购物车商品“规格/属性”。
    *   修改购物车商品“规格/属性”。
    *   修改购物车某个商品的数量。
    *   删除桌台订单商品。
    *   桌台订单商品改价。
    *   送厨购物车商品。
    *   退菜购物车商品。
    *   取消退菜购物车商品。
    *   转菜购物车商品。
    *   打包单商品。
    *   取消打包单商品。
    *   赠菜购物车商品。
    *   取消赠菜购物车商品。
*   **订单优惠与备注**:
    *   桌台订单打折 (支持改价、打折、抹零)。
    *   取消桌台订单所有优惠折扣。
    *   桌台订单商品备注。
    *   桌台订单整单备注。
    *   获取整单备注列表。
*   **自助餐相关操作**:
    *   桌台订单调整自助餐。
    *   桌台订单自助餐加钟。
    *   获取自助餐商品列表。
*   **结账流程**:
    *   确认必点商品。
    *   订单检查 (检查是否可以结账)。
    *   获取结账页面信息。
    *   选择或取消优惠券。
    *   设置订单的抵扣积分数量。
    *   获取支付方式的二维码信息。
    *   创建一个支付单。
    *   撤销一个支付单。
    *   完成销售订单的付款结账。
    *   免单。
    *   设置结账抹零规则。
*   **销售订单管理**:
    *   创建一个销售订单。
    *   从一个销售订单移动商品到另一个销售订单 (拆单)。
    *   删除一个销售订单 (删除拆单)。
    *   删除所有子销售订单 (撤销拆单)。
*   **会员管理**:
    *   获取订单会员优惠。
    *   确认使用会员优惠并验证密码。
    *   不使用此会员。
    *   获取订单会员列表。
*   **打印功能 (桌台)**:
    *   打印小票。
    *   打印发票。
*   **其他**:
    *   获取每日销售出库汇总。
    *   获取总部物品列表 (与 ERP 集成)。
    *   获取分批送厨弹框的销售订单商品列表。
    *   分批送厨。

#### 2.8 H5 订单管理 (由 `cashier_h5_order.go` 提供)

**概述**: 处理来自 H5 端的订单，包括订单列表、详情、接单和拒单功能。

**详细功能点**:

*   **H5 订单查询**:
    *   获取 H5 订单列表 (支持分页、状态筛选、区域筛选)。
    *   获取 H5 订单详情。
*   **H5 订单操作**:
    *   拒单。
    *   接单。

#### 2.9 即时订单（点餐）管理 (由 `cashier_instant.go` 提供)

**概述**: 处理收银端的即时订单，主要用于外带、快餐等不需要分配桌台的业务场景。该模块与桌台管理模块在订单商品操作、优惠、结账、会员等方面有大量功能重叠。

**详细功能点**:

*   **订单创建与管理**:
    *   创建即时订单 (点餐订单)。
    *   取消即时订单。
    *   隐藏即时订单 (挂单)。
    *   显示即时订单 (取单)。
    *   获取即时订单列表 (取单列表)。
    *   打包。
*   **购物车与订单商品操作** (与 `cashier_desk.go` 功能类似):
    *   查询即时订单购物车信息。
    *   向购物车添加商品。
    *   向购物车添加套餐。
    *   查询购物车商品“规格/属性”。
    *   修改购物车商品“规格/属性”。
    *   修改购物车某个商品的数量。
    *   删除即时订单商品。
    *   即时订单商品改价。
    *   送厨购物车商品。
    *   退菜购物车商品。
    *   取消退菜购物车商品。
    *   赠菜购物车商品。
    *   取消赠菜购物车商品。
*   **订单优惠与备注** (与 `cashier_desk.go` 功能类似):
    *   即时订单打折 (支持改价、打折、抹零)。
    *   取消即时订单所有优惠折扣。
    *   即时订单商品备注。
    *   即时订单整单备注。
    *   获取整单备注列表。
*   **结账流程** (与 `cashier_desk.go` 功能类似):
    *   确认必点商品。
    *   订单检查 (检查是否可以结账)。
    *   获取结账页面信息。
    *   选择或取消优惠券。
    *   设置订单的抵扣积分数量。
    *   获取支付方式的二维码信息。
    *   创建一个支付单。
    *   撤销一个支付单。
    *   完成销售订单的付款结账。
    *   免单。
    *   设置结账抹零规则。
*   **销售订单管理** (与 `cashier_desk.go` 功能类似):
    *   创建一个销售订单。
    *   从一个销售订单移动商品到另一个销售订单 (拆单)。
    *   删除一个销售订单 (删除拆单)。
    *   删除所有子销售订单 (撤销拆单)。
*   **会员管理** (与 `cashier_desk.go` 功能类似):
    *   获取订单会员优惠。
    *   确认使用会员优惠并验证密码。
    *   不使用此会员。
    *   获取订单会员列表。
*   **打印功能 (即时订单)**:
    *   打印小票。
    *   打印发票。
    *   订单解锁。
*   **其他**:
    *   获取分批送厨弹框的销售订单商品列表。
    *   分批送厨。

#### 2.10 会员外送订单接单与管理 (由 `cashier_member_order.go` 和 `cashier_member_order_manage.go` 提供)

**概述**: 处理会员外送订单的接单、拒单、备餐，以及订单列表、详情、搜索和退款管理。

**详细功能点**:

*   **订单接单流程**:
    *   获取外送订单接单列表 (支持分页、状态筛选)。
    *   获取外送订单详情。
    *   接单 (接受会员外送订单)。
    *   拒单 (拒绝会员外送订单)。
    *   备餐完成 (标记订单备餐完成)。
    *   取消订单。
*   **订单查询与搜索**:
    *   搜索订单列表 (通过关键词)。
*   **订单管理**:
    *   获取外送订单管理页面订单列表 (支持分页、状态筛选、订单编号、订单序号、日期和时间类型筛选)。
    *   获取外送订单管理页面订单详情。
*   **退款管理**:
    *   获取外送订单退款弹窗信息。
    *   外送订单退款/部分退款。
    *   外送订单重新退款。

#### 2.11 会员管理与充值订单 (由 `cashier_member.go` 提供)

**概述**: 提供收银端会员等级、卡类型查询，会员添加，会员信息充值，以及充值订单的创建、支付、撤销和打印功能。

**详细功能点**:

*   **会员信息**:
    *   获取会员等级列表。
    *   获取会员卡类型列表。
    *   模糊搜索会员。
    *   获取充值会员信息。
    *   添加会员。
*   **充值订单**:
    *   获取进行中的充值订单。
    *   创建充值订单。
    *   充值订单添加支付方式。
    *   充值订单撤销支付方式。
    *   确认充值订单。
    *   打印充值订单。
    *   获取充值订单支付二维码。
*   **会员密码验证**:
    *   使用会员优惠验证密码。

#### 2.12 历史订单管理 (由 `cashier_order_old.go` 提供)

**概述**: 处理收银端的历史订单查询、取消、删除、打印和退款功能。**该模块通过调用旧的 PHP 后端接口实现大部分功能。**

**详细功能点**:

*   **历史订单查询**:
    *   获取历史订单列表 (支持分页、状态、账单类型、日期类型、时间范围、订单号、订单类型筛选)。
    *   获取历史订单详情。
*   **历史订单操作**:
    *   取消历史订单。
    *   删除历史订单。
    *   获取历史订单退款信息。
    *   历史订单退款。
*   **打印与发票**:
    *   打印历史订单小票。
    *   打印历史订单发票。
    *   获取历史订单发票信息。
*   **余额处理**:
    *   处理会员余额 (旧订单退款场景)。
    *   处理钱箱余额 (旧订单退款场景)。

#### 2.13 打印记录管理 (由 `cashier_printer.go` 提供)

**概述**: 管理收银端的打印记录，包括查询打印配置、打印列表、以及打印和打印报告功能。

**详细功能点**:

*   **打印配置与列表**:
    *   获取打印列表查询条件。
    *   获取打印列表。
*   **打印操作**:
    *   打印 (根据请求参数打印)。
    *   打印报告。

#### 2.14 产品管理 (由 `cashier_product.go` 提供)

**概述**: 提供收银端产品和产品类别的查询功能。

**详细功能点**:

*   **产品查询**:
    *   获取收银产品列表 (支持分页)。
    *   获取收银产品类别列表。

#### 2.15 售罄管理 (由 `cashier_sold_out.go` 提供)

**概述**: 管理收银端商品的售罄状态，包括查询售罄商品、设置和取消售罄。

**详细功能点**:

*   **售罄商品查询**:
    *   获取售罄商品列表。
*   **售罄状态管理**:
    *   设置商品售罄。
    *   取消商品售罄。

#### 2.16 统计管理 (由 `cashier_statistics.go` 提供)

**概述**: 提供收银端统计报表功能，包括销售统计、支付方式统计等。

**详细功能点**:

*   **销售统计**:
    *   获取销售汇总统计。
    *   获取销售明细列表。
    *   获取销售趋势图数据。
*   **支付方式统计**:
    *   获取支付方式统计。

### 3. 数据模型 (DTO - Data Transfer Objects)

收银端功能与以下 DTO 结构体密切相关 (示例，具体结构体需查阅 `main/app/dto/req` 和 `main/app/dto/resp`):

*   **请求 (Request)**: `req.LoginReq`, `req.OrderListReq`, `req.OrderInfoReq`, `req.OrderCancelReq`, `req.VerifyPasswordReq`, `req.UpdateAcceptOrderSetting`, `req.ShiftWithdrawReq`, `req.UsbPrinterReportReq`, `req.DecryptQrCodeReq`, `req.DeskListReq`, `req.DeskInfoReq`, `req.DeskOrderCreateReq`, `req.OrderCartProductAddReq`, `req.InstantOrderPaymentFinishReq`, `req.H5OrderListReq`, `req.H5OrderDetailReq`, `req.OrderTakeoutReq`, `req.InstantOrderSaleOrderCreateReq`, `req.MemberOrderListReq`, `req.GetMemberOrderDetailReq`, `req.AcceptOrderReq`, `req.RejectOrderReq`, `req.CookFinishOrderReq`, `req.MemberOrderSearchReq`, `member_req.CancelOrderReq`, `req.MemberOrderManageListReq`, `req.GetMemberOrderManageDetailReq`, `member_req.MemberOrderReturnInfoReq`, `req.OrderReturnReq`, `req.OrderReReturnReq`, `req.AddMemberReq`, `req.RechargeReq`, `req.RechargeOrderAddPaymentMethodReq`, `req.RechargeOrderCancelPaymentMethodReq`, `req.ConfirmRechargeOrder`, `req.PrintRechargeOrderReq`, `req.CheckMemberPasswordReq`, `req.RechargeOrderPaymentQrcodeReq`, `req.MemberBalanceChangeReq`, `req.CashBoxBalanceChangeReq`, `req.PrinterListReq`, `req.PrinterPrintReq`, `req.PrinterReportReqs`, `req.ProductListReq`, `req.SoldOutProductListReq`, `req.SetSoldOutReq`, `req.CancelSoldOutReq`, `req.SalesStatisticsReq`, `req.PaymentMethodStatisticsReq` 等。
*   **响应 (Response)**: `resp.CashierLoginResp`, `resp.CashierBase`, `resp.LanguageResp`, `resp.Ads`, `resp.PaymentMethodList`, `resp.BuffetListPaginationResp`, `resp.BuffetDelayListResp`, `resp.AbnormalPrintList`, `resp.UnprocessedCallList`, `resp.UnprocessedResp`, `resp.OrderListPaginationResp`, `resp.OrderInfosResp`, `resp.DeskRegionAndTypeListWithPaginationResp`, `resp.DeskListWithPaginationResp`, `resp.Desk`, `resp.CreateDeskOrderResp`, `resp.ShopCart`, `resp.InstantOrderHideOrderListResp`, `resp.InstantOrderPaymentInfoResp`, `resp.OrderFinishResp`, `resp.H5OrderList`, `resp.H5OrderDetailResp`, `resp.OrderCheckRes`, `resp.GetMemberCashierOrderListResp`, `resp.GetMemberOrderDetailResp`, `resp.GetMemberCashierOrderSearchResp`, `resp.GetMemberOrderManageListResp`, `resp.GetMemberOrderManageDetailResp`, `resp.MemberLevelList`, `resp.MemberCardTypeList`, `resp.SearchMemberList`, `resp.RechargeMember`, `resp.RechargeOrder`, `resp.ConfirmRechargeOrder`, `resp.RechargeOrderPaymentQrcodeInfoResp`, `resp.SaleOrderInvoiceInfo`, `resp.PrinterLogData`, `resp.PrinterBaseResp`, `resp.PrinterListPaginationResp`, `resp.PrinterData`, `product_resp.ProductListWithPaginationResp`, `product_resp.ProductCategoryListResp`, `resp.SoldOutProductListResp`, `resp.SalesSummaryStatistics`, `resp.SalesDetailList`, `resp.SalesTrendGraph`, `resp.PaymentMethodStatistics` 等。

### 4. 服务依赖

收银端控制器依赖以下服务层接口 (由 `main/app/service` 和 `main/app/printer/service` 提供):

*   `service.IAuthSrv`: 认证服务，用于收银员登录、刷新令牌、退出登录，以及获取收银员基本信息和权限验证。
*   `setting.ISrv`: 系统设置服务，用于获取和修改各项系统配置。
*   `service.IPaymentMethodSrv`: 支付方式服务，用于获取支付方式列表。
*   `service.IOtherSrv`: 其他服务，用于获取退菜/免单/赠菜原因、整单备注列表。
*   `printerService.IPrinterLogSrv`: 打印日志服务，用于获取打印数据、重新打印，以及打印记录管理。
*   `service.IStaffShiftSrv`: 员工交班服务，用于交班管理和报备。
*   `service.IPrinterSrv`: 打印服务，用于 USB 打印机上报。
*   `service.IMarketingActivitySrv`: 营销活动服务，用于解密活动二维码。
*   `service.IBuffetSrv`: 自助餐服务，用于获取自助餐列表和延迟列表。
*   `service.ICallSrv`: 呼叫服务，用于处理异常打印和未处理呼叫。
*   `service.IOrderSrv`: 订单服务，负责核心订单业务逻辑，包括购物车操作、订单优惠、结账流程、销售订单管理、会员优惠等。
*   `service.IDeskSrv`: 桌台服务，用于桌台的生命周期管理、状态判断等。
*   `service.IMemberSrv`: 会员服务，用于获取会员等级、卡类型、搜索会员、获取充值会员信息、添加会员、会员密码验证，以及处理会员余额 (旧订单退款场景)。
*   `service.IH5OrderSrv`: H5 订单服务，用于 H5 订单的列表、详情、接单和拒单。
*   `service.IMemberOrderSrv`: 会员订单服务，用于会员外送订单的接单、拒单、备餐完成、订单查询、搜索和退款管理。
*   `service.IRechargeOrderSrv`: 充值订单服务，用于充值订单的创建、支付、撤销、确认、打印和获取支付二维码。
*   `service.ICashBoxSrv`: 钱箱服务，用于处理钱箱余额 (旧订单退款场景)。
*   `service.IProductSrv`: 产品服务，用于获取收银产品列表和产品类别列表。
*   `service.ISoldOutSrv`: 售罄服务，用于管理商品售罄状态。
*   `service.IStatisticsSrv`: 统计服务，用于获取销售和支付方式统计数据。

### 5. 技术栈与框架

*   **编程语言**: Go
*   **Web框架**: Gin
*   **依赖管理**: Go Modules
*   **数据库**: GORM (通过 `pkg/database.DBManager` 管理)
*   **缓存**: (通过 `pkg/cache.Cache` 管理)
*   **认证**: JWT Token (`middleware.Auth`)
*   **错误处理**: 自定义错误 (`ttpos-server-go/app/errors`)
*   **国际化**: `ttpos-server-go/i18n`
*   **API文档**: Swagger (通过注释自动生成)
*   **日志**: `go.uber.org/zap`
*   **内部 HTTP 调用**: `ttpos-server-go/pkg/utils.HttpPost`, `ttpos-server-go/pkg/utils.HttpGet` (主要用于 `cashier_order_old.go` 调用旧的 PHP 后端接口)

### 6. 接口规范

*   遵循 RESTful API 设计原则。
*   URL 使用 Snake Case (例如 `/cashier/verify_cash_box_password`)。
*   统一响应格式 (`dto.Response` 结构)。
*   请求参数通过 `gin.Context.ShouldBindJSON` 或 `gin.Context.ShouldBindQuery` 进行绑定和验证。
*   错误处理统一使用 `helper.ErrorWithDetail` 或 `helper.Fail`。
*   成功响应统一使用 `helper.Success`。

### 7. 异常处理

*   所有业务逻辑错误通过 `errors.WithMessage(err)` 进行包装，以便追踪错误栈。
*   参数验证错误通过 `helper.HandleValidationError` 统一处理。
*   不使用 `panic` 处理业务逻辑错误，而是返回 `error`。

### 8. 性能与安全

*   对敏感操作 (如钱箱密码验证) 增加了密码验证。
*   API 响应时间有要求 (本地响应时间要求 200ms 以内)。
*   接口访问需要 JWT Token 认证。
*   对库存不足等情况进行了错误处理和提示。
*   对内部接口调用进行了 IP 限制 (例如 `HandleMemberBalance` 和 `HandleCashBoxBalance` 只能在内网环境下访问)。

---
