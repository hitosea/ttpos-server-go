# 功能实现设计文档：收银端核心功能

### 1. 简介

本文档旨在详细描述 TTPOS-Server-Go 项目中收银端（Cashier）模块各个核心功能在代码层面的实现设计。我们将深入探讨每个功能模块涉及的控制器、服务、数据访问层、数据传输对象以及相关的技术栈和框架应用。

### 2. 功能模块实现设计

#### 2.1 基础信息与系统设置

**文件**: `main/app/api/v1/cashier/cashier_base.go`

**概述**: 此模块提供收银端运行所需的基础信息、系统配置管理以及相关的安全验证功能。

**控制器层 (`cashier_base.go`) 实现**:

*   **API 接口**:
    *   `/api/v1/cashier/base_info`: 获取收银端基本信息。
    *   `/api/v1/cashier/language_config`: 获取多语言配置。
    *   `/api/v1/cashier/ads`: 获取副屏广告信息。
    *   `/api/v1/cashier/verify_cash_box_password`: 验证钱箱密码。
    *   `/api/v1/cashier/verify_advance_password`: 验证高级密码。
    *   `/api/v1/cashier/verify_screen_lock_password`: 验证锁屏密码。
    *   `/api/v1/cashier/version_update_check`: 检查收银端版本更新。
    *   `/api/v1/cashier/payment_method_list`: 获取支付方式列表。
    *   `/api/v1/cashier/update_accept_order_setting`: 修改接单设置。
    *   `/api/v1/cashier/update_member_order_accept_setting`: 修改会员订单接单设置。
    *   `/api/v1/cashier/update_system_common_setting`: 修改系统通用设置。
    *   `/api/v1/cashier/setting_info`: 获取收银端所有设置信息。
    *   `/api/v1/cashier/refund_product_reason_list`: 获取退菜原因列表。
    *   `/api/v1/cashier/free_product_reason_list`: 获取免单/赠菜原因列表。
    *   `/api/v1/cashier/get_wait_print_data`: 获取待打印数据。
    *   `/api/v1/cashier/get_open_cash_box_printer_config`: 获取打开钱箱的打印机配置。
    *   `/api/v1/cashier/usb_printer_report`: USB打印机上报功能。
    *   `/api/v1/cashier/decrypt_activity_qr_code`: 解密活动二维码。
*   **路由处理**: 使用 Gin 框架的 `gin.Context` 处理请求和响应。通过 `ShouldBindJSON` 或 `ShouldBindQuery` 绑定请求参数到 DTO 结构体。
*   **错误处理**: 使用 `helper.ErrorWithDetail` 或 `helper.Fail` 返回错误信息，`helper.Success` 返回成功响应。
*   **服务调用**: 调用 `service.IAuthSrv`, `setting.ISrv`, `service.IPaymentMethodSrv`, `service.IOtherSrv`, `printerService.IPrinterLogSrv`, `service.IPrinterSrv`, `service.IMarketingActivitySrv` 等服务接口处理业务逻辑。

**服务层实现**:

*   **`service.IAuthSrv`**: 包含验证密码、获取收银员基本信息和权限验证等方法。
*   **`setting.ISrv`**: 包含获取和修改系统各项配置的方法。
*   **`service.IPaymentMethodSrv`**: 包含获取支付方式列表的方法。
*   **`service.IOtherSrv`**: 包含获取退菜/免单/赠菜原因、整单备注列表等方法。
*   **`printerService.IPrinterLogSrv`**: 包含获取待打印数据、重新打印等方法。
*   **`service.IPrinterSrv`**: 包含 USB 打印机上报等方法。
*   **`service.IMarketingActivitySrv`**: 包含解密活动二维码等方法。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.VerifyPasswordReq`, `req.UpdateAcceptOrderSetting`, `req.UsbPrinterReportReq`, `req.DecryptQrCodeReq` 等。
*   **响应 (Response)**: `resp.CashierBase`, `resp.LanguageResp`, `resp.Ads`, `resp.PaymentMethodList` 等。

**其他技术栈应用**:

*   **认证鉴权**: API 接口通过 `middleware.Auth` 进行 JWT Token 认证。
*   **错误处理**: 统一使用 `ttpos-server-go/app/errors` 进行错误包装。
*   **国际化**: 可能使用 `ttpos-server-go/i18n` 获取多语言配置。
*   **日志**: 通过 `go.uber.org/zap` 记录日志。

#### 2.2 交班管理

**文件**: `main/app/api/v1/cashier/cashier_base.go` (部分功能)

**概述**: 支持收银员进行班次交接，包括交班信息的查询、提交、现金操作以及交班单据的打印。

**控制器层 (`cashier_base.go`) 实现**:

*   **API 接口**:
    *   `/api/v1/cashier/shift_info`: 获取当前班次交班信息。
    *   `/api/v1/cashier/shift_submit`: 提交交班。
    *   `/api/v1/cashier/shift_withdraw`: 交班取钱操作。
    *   `/api/v1/cashier/shift_deposit`: 交班存钱操作。
    *   `/api/v1/cashier/print_shift_report`: 打印交班单据。
    *   `/api/v1/cashier/report_info`: 获取报备信息。
    *   `/api/v1/cashier/report_submit`: 提交报备信息。
*   **服务调用**: 调用 `service.IStaffShiftSrv` 处理交班和报备的业务逻辑。

**服务层实现**:

*   **`service.IStaffShiftSrv`**: 包含获取交班信息、提交交班、取钱、存钱、打印交班单据、获取报备信息、提交报备信息等方法。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.ShiftWithdrawReq`, `req.ShiftDepositReq`, `req.ReportSubmitReq` 等。
*   **响应 (Response)**: `resp.ShiftInfo`, `resp.ReportInfo` 等。

**其他技术栈应用**:

*   **并发控制**: 在处理交班取钱、存钱等操作时，可能会使用 `pkg/lock.Lock` 进行并发控制，确保数据一致性。
*   **打印**: 调用打印服务生成交班单据。

#### 2.3 订单管理

**文件**: `main/app/api/v1/cashier/cashier_order.go`

**概述**: 负责收银端订单的创建、查询、修改、取消、退款、反结账、删除以及打印功能。

**控制器层 (`cashier_order.go`) 实现**:

*   **API 接口**:
    *   `/api/v1/cashier/order/list`: 获取订单列表。
    *   `/api/v1/cashier/order/info`: 获取订单详情。
    *   `/api/v1/cashier/order/cancel`: 取消订单。
    *   `/api/v1/cashier/order/refund_info`: 获取退款信息。
    *   `/api/v1/cashier/order/refund`: 退款订单。
    *   `/api/v1/cashier/order/re_refund`: 重新退款。
    *   `/api/v1/cashier/order/anti_settle_info`: 获取反结账弹窗信息。
    *   `/api/v1/cashier/order/anti_settle`: 反结账。
    *   `/api/v1/cashier/order/delete`: 删除订单。
    *   `/api/v1/cashier/order/can_close`: 判断订单是否可关闭。
    *   `/api/v1/cashier/order/print_receipt`: 打印小票。
    *   `/api/v1/cashier/order/print_invoice`: 打印发票。
    *   `/api/v1/cashier/order/invoice_info`: 获取发票信息。
*   **服务调用**: 调用 `service.IOrderSrv` 处理订单业务逻辑。

**服务层实现**:

*   **`service.IOrderSrv`**: 包含订单列表查询、订单详情获取、取消订单、退款、反结账、删除订单、判断订单是否可关闭、打印小票/发票、获取发票信息等方法。
*   **`service.IPaymentMethodSrv`**: 可能在退款或结账时需要获取支付方式信息。
*   **`printerService.IPrinterLogSrv`**: 用于记录打印日志。

**数据访问层 (Repository) 可能涉及**:

*   **`repository.IOrderRepo`**: 订单数据访问，包含订单的 CRUD 操作，以及复杂的查询，如按状态、时间范围、用户等筛选。
*   **`repository.IOrderItemRepo`**: 订单商品数据访问。
*   **`repository.IPaymentRepo`**: 支付单据数据访问。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.OrderListReq`, `req.OrderInfoReq`, `req.OrderCancelReq`, `req.OrderRefundReq` 等。
*   **响应 (Response)**: `resp.OrderListPaginationResp`, `resp.OrderInfosResp` 等。

**其他技术栈应用**:

*   **数据库操作**: 通过 `database.DBManager` 获取数据库连接，使用 GORM 进行数据操作。
*   **事务管理**: 在取消、退款、反结账等复杂操作中，通常需要使用数据库事务来保证数据一致性。
*   **事件总线**: 订单状态变更可能会发布事件，例如 `EventOrderCreated`, `EventOrderRefunded` 等，由其他模块订阅处理。
*   **并发控制**: 在修改订单状态时，可能会使用并发锁。

#### 2.4 认证与鉴权

**文件**: `main/app/api/v1/cashier/cashier_auth.go`

**概述**: 处理收银员的登录、刷新令牌和退出登录等认证相关功能。

**控制器层 (`cashier_auth.go`) 实现**:

*   **API 接口**:
    *   `/api/v1/cashier/login`: 收银员登录。
    *   `/api/v1/cashier/refresh_token`: 刷新访问令牌。
    *   `/api/v1/cashier/logout`: 退出登录。
*   **服务调用**: 调用 `service.IAuthSrv` 处理认证和鉴权逻辑。

**服务层实现**:

*   **`service.IAuthSrv`**: 包含用户登录验证、生成 JWT Token、刷新 Token、使 Token 失效等方法。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.LoginReq`。
*   **响应 (Response)**: `resp.CashierLoginResp` (包含 Token 和 RefreshToken)。

**其他技术栈应用**:

*   **JWT**: 使用 JWT (JSON Web Tokens) 进行身份验证和授权。Token 的生成、解析、验证和刷新是核心。
*   **密码加密**: 用户密码存储时会进行加密处理 (例如 bcrypt)。
*   **Redis/缓存**: Refresh Token 可能存储在 Redis 等缓存中，用于管理 Token 的生命周期和黑名单。

#### 2.5 自助餐管理

**文件**: `main/app/api/v1/cashier/cashier_buffet.go`

**概述**: 提供收银端自助餐相关的查询功能。

**控制器层 (`cashier_buffet.go`) 实现**:

*   **API 接口**:
    *   `/api/v1/cashier/buffet_list`: 获取所有自助餐方案的列表。
    *   `/api/v1/cashier/buffet_delay_list`: 获取自助餐延迟的列表。
*   **服务调用**: 调用 `service.IBuffetSrv` 处理自助餐业务逻辑。

**服务层实现**:

*   **`service.IBuffetSrv`**: 包含获取自助餐方案列表、获取自助餐延迟列表等方法。

**数据访问层 (Repository) 可能涉及**:

*   **`repository.IBuffetRepo`**: 自助餐方案数据访问。
*   **`repository.IBuffetDelayRepo`**: 自助餐延迟数据访问。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: (可能为空或包含分页参数)。
*   **响应 (Response)**: `resp.BuffetListPaginationResp`, `resp.BuffetDelayListResp`。

#### 2.6 呼叫管理

**文件**: `main/app/api/v1/cashier/cashier_call.go`

**概述**: 处理收银端接收到的呼叫信息，包括异常打印和未处理呼叫，并支持相关操作。

**控制器层 (`cashier_call.go`) 实现**:

*   **API 接口**:
    *   `/api/v1/cashier/call/abnormal_print_list`: 获取异常打印列表。
    *   `/api/v1/cashier/call/delete_abnormal_print`: 删除异常打印记录。
    *   `/api/v1/cashier/call/re_print`: 重新打印。
    *   `/api/v1/cashier/call/unprocessed_list`: 获取未处理呼叫列表。
    *   `/api/v1/cashier/call/unprocessed_count`: 获取未处理呼叫的数量。
    *   `/api/v1/cashier/call/unprocessed_notify`: 获取未处理呼叫的通知信息。
    *   `/api/v1/cashier/call/handle`: 处理呼叫。
*   **服务调用**: 调用 `service.ICallSrv` 处理呼叫业务逻辑。
*   **服务调用**: 调用 `printerService.IPrinterLogSrv` 处理异常打印的重新打印。

**服务层实现**:

*   **`service.ICallSrv`**: 包含获取异常打印列表、删除异常打印记录、获取未处理呼叫列表、处理呼叫等方法。
*   **`printerService.IPrinterLogSrv`**: 包含重新打印等方法。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.DeleteAbnormalPrintReq`, `req.HandleCallReq` 等。
*   **响应 (Response)**: `resp.AbnormalPrintList`, `resp.UnprocessedCallList`, `resp.UnprocessedResp` 等。

**其他技术栈应用**:

*   **Websocket**: 呼叫功能可能通过 WebSocket 进行实时推送。

#### 2.7 桌台管理

**文件**: `main/app/api/v1/cashier/cashier_desk.go`

**概述**: 收银端桌台管理的核心模块，涵盖了桌台的生命周期管理、订单商品的精细化操作、结账流程、会员优惠以及与 ERP 系统和打印相关的集成。

**控制器层 (`cashier_desk.go`) 实现**:

*   **API 接口**:
    *   `/api/v1/cashier/desk/region_type_list`: 获取桌台区域和类型列表。
    *   `/api/v1/cashier/desk/list`: 获取桌台列表。
    *   `/api/v1/cashier/desk/info`: 获取桌台详情。
    *   `/api/v1/cashier/desk/open`: 开台。
    *   `/api/v1/cashier/desk/close`: 关闭桌台。
    *   `/api/v1/cashier/desk/clear`: 清台。
    *   `/api/v1/cashier/desk/switch`: 切换桌台。
    *   `/api/v1/cashier/desk/merge`: 合并桌台。
    *   `/api/v1/cashier/desk/cancel_order`: 取消桌台订单。
    *   `/api/v1/cashier/desk/order_unlock`: 订单解锁。
    *   `/api/v1/cashier/desk/shop_cart`: 查询桌台购物车信息。
    *   `/api/v1/cashier/desk/order_product_add`: 向购物车添加商品。
    *   `/api/v1/cashier/desk/order_package_add`: 向购物车添加套餐。
    *   `/api/v1/cashier/desk/product_spec_attr_info`: 查询购物车商品“规格/属性”。
    *   `/api/v1/cashier/desk/update_product_spec_attr`: 修改购物车商品“规格/属性”。
    *   `/api/v1/cashier/desk/update_product_num`: 修改购物车某个商品的数量。
    *   `/api/v1/cashier/desk/delete_product`: 删除桌台订单商品。
    *   `/api/v1/cashier/desk/update_product_price`: 桌台订单商品改价。
    *   `/api/v1/cashier/desk/cook_product`: 送厨购物车商品。
    *   `/api/v1/cashier/desk/refund_product`: 退菜购物车商品。
    *   `/api/v1/cashier/desk/cancel_refund_product`: 取消退菜购物车商品。
    *   `/api/v1/cashier/desk/transfer_product`: 转菜购物车商品。
    *   `/api/v1/cashier/desk/packing_product`: 打包单商品。
    *   `/api/v1/cashier/desk/cancel_packing_product`: 取消打包单商品。
    *   `/api/v1/cashier/desk/free_product`: 赠菜购物车商品。
    *   `/api/v1/cashier/desk/cancel_free_product`: 取消赠菜购物车商品。
    *   `/api/v1/cashier/desk/discount_order`: 桌台订单打折。
    *   `/api/v1/cashier/desk/cancel_all_discount`: 取消桌台订单所有优惠折扣。
    *   `/api/v1/cashier/desk/product_remark`: 桌台订单商品备注。
    *   `/api/v1/cashier/desk/order_remark`: 桌台订单整单备注。
    *   `/api/v1/cashier/desk/order_remark_list`: 获取整单备注列表。
    *   `/api/v1/cashier/desk/buffet_adjust`: 桌台订单调整自助餐。
    *   `/api/v1/cashier/desk/buffet_add_time`: 桌台订单自助餐加钟。
    *   `/api/v1/cashier/desk/buffet_product_list`: 获取自助餐商品列表。
    *   `/api/v1/cashier/desk/check_must_product`: 确认必点商品。
    *   `/api/v1/cashier/desk/order_check`: 订单检查。
    *   `/api/v1/cashier/desk/payment_page_info`: 获取结账页面信息。
    *   `/api/v1/cashier/desk/select_coupon`: 选择或取消优惠券。
    *   `/api/v1/cashier/desk/set_deduct_points`: 设置订单的抵扣积分数量。
    *   `/api/v1/cashier/desk/payment_qrcode_info`: 获取支付方式的二维码信息。
    *   `/api/v1/cashier/desk/create_payment`: 创建一个支付单。
    *   `/api/v1/cashier/desk/cancel_payment`: 撤销一个支付单。
    *   `/api/v1/cashier/desk/payment_finish`: 完成销售订单的付款结账。
    *   `/api/v1/cashier/desk/free_order`: 免单。
    *   `/api/v1/cashier/desk/set_chop_zero_rule`: 设置结账抹零规则。
    *   `/api/v1/cashier/desk/create_sale_order`: 创建一个销售订单。
    *   `/api/v1/cashier/desk/move_product_to_sale_order`: 从一个销售订单移动商品到另一个销售订单。
    *   `/api/v1/cashier/desk/delete_sale_order`: 删除一个销售订单。
    *   `/api/v1/cashier/desk/delete_all_child_sale_order`: 删除所有子销售订单。
    *   `/api/v1/cashier/desk/member_order_discount`: 获取订单会员优惠。
    *   `/api/v1/cashier/desk/confirm_member_discount`: 确认使用会员优惠并验证密码。
    *   `/api/v1/cashier/desk/cancel_member`: 不使用此会员。
    *   `/api/v1/cashier/desk/member_list`: 获取订单会员列表。
    *   `/api/v1/cashier/desk/print_receipt`: 打印小票。
    *   `/api/v1/cashier/desk/print_invoice`: 打印发票。
    *   `/api/v1/cashier/desk/daily_sale_out_summary`: 获取每日销售出库汇总。
    *   `/api/v1/cashier/desk/head_office_product_list`: 获取总部物品列表。
    *   `/api/v1/cashier/desk/batch_cook_sale_order_product_list`: 获取分批送厨弹框的销售订单商品列表。
    *   `/api/v1/cashier/desk/batch_cook`: 分批送厨。
*   **服务调用**: 调用 `service.IDeskSrv`, `service.IOrderSrv`, `service.IMemberSrv`, `service.IPaymentMethodSrv`, `service.IOtherSrv`, `service.IBuffetSrv`, `printerService.IPrinterLogSrv`, `service.IProductSrv`, `service.IAuthSrv`, `service.IStockOutSrv` (可能涉及每日销售出库汇总) 等服务处理业务逻辑。

**服务层实现**:

*   **`service.IDeskSrv`**: 负责桌台的生命周期管理 (开台、关台、清台、转台、合并、取消订单、订单解锁)、桌台信息查询等。
*   **`service.IOrderSrv`**: 负责购物车操作 (添加、修改、删除商品、规格属性、数量)、订单优惠 (打折、抹零、取消折扣)、商品备注、整单备注、自助餐调整和加钟、订单检查、结账流程 (创建/撤销支付单、付款结账、免单、抹零规则)、销售订单管理 (创建、拆单、删除)、必点商品确认、会员优惠处理 (确认/取消会员、获取会员列表)、打印功能等。
*   **`service.IMemberSrv`**: 负责获取订单会员优惠、确认使用会员优惠并验证密码、不使用此会员、获取订单会员列表。
*   **`service.IPaymentMethodSrv`**: 负责获取支付方式的二维码信息。
*   **`service.IOtherSrv`**: 负责获取整单备注列表。
*   **`service.IBuffetSrv`**: 负责桌台订单调整自助餐、自助餐加钟、获取自助餐商品列表。
*   **`printerService.IPrinterLogSrv`**: 负责打印小票、发票。
*   **`service.IProductSrv`**: 负责获取总部物品列表。
*   **`service.IAuthSrv`**: 负责验证密码。
*   **`service.IStockOutSrv`**: 负责获取每日销售出库汇总。

**数据访问层 (Repository) 可能涉及**:

*   **`repository.IDeskRepo`**: 桌台数据访问。
*   **`repository.IOrderRepo`**: 订单数据访问。
*   **`repository.IOrderItemRepo`**: 订单商品数据访问。
*   **`repository.IMemberRepo`**: 会员数据访问。
*   **`repository.ICouponRepo`**: 优惠券数据访问。
*   **`repository.IPaymentRepo`**: 支付单据数据访问。
*   **`repository.IProductRepo`**: 产品数据访问。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.DeskListReq`, `req.DeskInfoReq`, `req.DeskOrderCreateReq`, `req.OrderCartProductAddReq`, `req.InstantOrderPaymentFinishReq`, `req.H5OrderListReq`, `req.DeskMergeReq`, `req.DeskSwitchReq`, `req.OrderProductUpdateNumReq`, `req.OrderProductRemarkReq`, `req.OrderRemarkReq`, `req.DeskDiscountOrderReq`, `req.DeskSetDeductPointsReq`, `req.DeskCreatePaymentReq`, `req.DeskCancelPaymentReq`, `req.DeskSetChopZeroRuleReq`, `req.DeskCreateSaleOrderReq`, `req.DeskMoveProductToSaleOrderReq`, `req.DeskDeleteSaleOrderReq`, `req.DeskDeleteAllChildSaleOrderReq`, `req.DeskConfirmMemberDiscountReq`, `req.DeskMemberOrderDiscountReq`, `req.DeskBatchCookReq` 等。
*   **响应 (Response)**: `resp.DeskRegionAndTypeListWithPaginationResp`, `resp.DeskListWithPaginationResp`, `resp.Desk`, `resp.CreateDeskOrderResp`, `resp.ShopCart`, `resp.InstantOrderPaymentInfoResp`, `resp.OrderFinishResp`, `resp.OrderCheckRes` 等。

**其他技术栈应用**:

*   **数据库事务**: 结账、拆单、合并桌台等操作涉及多个表的更新，需要严格的事务管理。
*   **并发锁**: 桌台状态、订单状态、库存等关键资源在多用户操作时需要并发锁保证数据一致性。
*   **事件总线**: 订单商品送厨、订单完成等操作可能发布事件，触发厨房打印、库存扣减等后续流程。
*   **ERP 集成**: 每日销售出库汇总、获取总部物品列表等功能可能涉及与外部 ERP 系统的接口调用。

#### 2.8 H5 订单管理

**文件**: `main/app/api/v1/cashier/cashier_h5_order.go`

**概述**: 处理来自 H5 端的订单，包括订单列表、详情、接单和拒单功能。

**控制器层 (`cashier_h5_order.go`) 实现**:

*   **API 接口**:
    *   `/api/v1/cashier/h5_order_list`: 获取 H5 订单列表。
    *   `/api/v1/cashier/h5_order_info`: 获取 H5 订单详情。
    *   `/api/v1/cashier/h5_order_reject`: 拒单。
    *   `/api/v1/cashier/h5_order_accept`: 接单。
*   **服务调用**: 调用 `service.IH5OrderSrv` 处理 H5 订单业务逻辑。

**服务层实现**:

*   **`service.IH5OrderSrv`**: 包含获取 H5 订单列表、详情、拒单、接单等方法。

**数据访问层 (Repository) 可能涉及**:

*   **`repository.IH5OrderRepo`**: H5 订单数据访问。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.H5OrderListReq`, `req.H5OrderDetailReq`, `req.H5OrderRejectReq`, `req.H5OrderAcceptReq` 等。
*   **响应 (Response)**: `resp.H5OrderList`, `resp.H5OrderDetailResp` 等。

**其他技术栈应用**:

*   **事件总线**: H5 订单接单/拒单可能发布事件，通知用户或其他系统。

#### 2.9 即时订单（点餐）管理

**文件**: `main/app/api/v1/cashier/cashier_instant.go`

**概述**: 处理收银端的即时订单，主要用于外带、快餐等不需要分配桌台的业务场景。该模块与桌台管理模块在订单商品操作、优惠、结账、会员等方面有大量功能重叠。

**控制器层 (`cashier_instant.go`) 实现**:

*   **API 接口**: (与 `cashier_desk.go` 中大部分订单相关的接口类似，URL 前缀不同，例如 `/api/v1/cashier/instant_order/...`)
    *   `/api/v1/cashier/instant_order/create`: 创建即时订单。
    *   `/api/v1/cashier/instant_order/cancel`: 取消即时订单。
    *   `/api/v1/cashier/instant_order/hide`: 隐藏即时订单 (挂单)。
    *   `/api/v1/cashier/instant_order/show`: 显示即时订单 (取单)。
    *   `/api/v1/cashier/instant_order/list`: 获取即时订单列表 (取单列表)。
    *   `/api/v1/cashier/instant_order/packing`: 打包。
    *   购物车、订单商品、优惠、结账、会员、打印等操作接口与 `cashier_desk.go` 类似，只是作用于即时订单。
*   **服务调用**: 主要调用 `service.IOrderSrv` 处理即时订单的业务逻辑，也可能调用 `service.IPaymentMethodSrv`, `service.IMemberSrv` 等。

**服务层实现**:

*   **`service.IOrderSrv`**: 负责即时订单的创建、取消、挂单、取单、打包，以及与桌台订单类似的购物车、订单商品、优惠、结账、会员、打印等操作。

**数据访问层 (Repository) 可能涉及**:

*   与 `cashier_desk.go` 模块类似，涉及 `repository.IOrderRepo`, `repository.IOrderItemRepo`, `repository.IPaymentRepo`, `repository.IMemberRepo` 等。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.InstantOrderCreateReq`, `req.InstantOrderCancelReq`, `req.OrderTakeoutReq`, `req.InstantOrderPaymentFinishReq` 等。与 `cashier_desk.go` 相关的 DTO 结构体也会在这里复用。
*   **响应 (Response)**: `resp.InstantOrderHideOrderListResp`, `resp.InstantOrderPaymentInfoResp`, `resp.OrderFinishResp` 等。

**其他技术栈应用**:

*   **代码复用**: 由于即时订单和桌台订单在很多业务逻辑上是相似的，服务层和部分控制器逻辑会进行高度复用。

#### 2.10 会员外送订单接单与管理

**文件**: `main/app/api/v1/cashier/cashier_member_order.go` 和 `main/app/api/v1/cashier/cashier_member_order_manage.go`

**概述**: 处理会员外送订单的接单、拒单、备餐，以及订单列表、详情、搜索和退款管理。

**控制器层实现**:

*   **`cashier_member_order.go` API 接口**:
    *   `/api/v1/cashier/member_order/list`: 获取外送订单接单列表。
    *   `/api/v1/cashier/member_order/info`: 获取外送订单详情。
    *   `/api/v1/cashier/member_order/accept`: 接单。
    *   `/api/v1/cashier/member_order/reject`: 拒单。
    *   `/api/v1/cashier/member_order/cook_finish`: 备餐完成。
    *   `/api/v1/cashier/member_order/cancel`: 取消订单。
    *   `/api/v1/cashier/member_order/search`: 搜索订单列表。
*   **`cashier_member_order_manage.go` API 接口**:
    *   `/api/v1/cashier/member_order_manage/list`: 获取外送订单管理页面订单列表。
    *   `/api/v1/cashier/member_order_manage/info`: 获取外送订单管理页面订单详情。
    *   `/api/v1/cashier/member_order_manage/return_info`: 获取外送订单退款弹窗信息。
    *   `/api/v1/cashier/member_order_manage/return`: 外送订单退款/部分退款。
    *   `/api/v1/cashier/member_order_manage/re_return`: 外送订单重新退款。
*   **服务调用**: 调用 `service.IMemberOrderSrv` 处理会员外送订单的业务逻辑。

**服务层实现**:

*   **`service.IMemberOrderSrv`**: 包含获取外送订单接单列表、详情、接单、拒单、备餐完成、取消订单、搜索订单、获取管理页面订单列表和详情、获取退款信息、退款和重新退款等方法。
*   **`service.IOrderSrv`**: 可能会在内部调用，以执行订单相关的通用操作。
*   **`service.IPaymentMethodSrv`**: 处理退款时可能需要。

**数据访问层 (Repository) 可能涉及**:

*   **`repository.IMemberOrderRepo`**: 会员外送订单数据访问。
*   **`repository.IOrderRepo`**: 可能涉及底层订单数据。
*   **`repository.IPaymentRepo`**: 退款操作可能涉及支付单据。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.MemberOrderListReq`, `req.GetMemberOrderDetailReq`, `req.AcceptOrderReq`, `req.RejectOrderReq`, `req.CookFinishOrderReq`, `req.MemberOrderSearchReq`, `member_req.CancelOrderReq`, `req.MemberOrderManageListReq`, `req.GetMemberOrderManageDetailReq`, `member_req.MemberOrderReturnInfoReq`, `req.OrderReturnReq`, `req.OrderReReturnReq` 等。
*   **响应 (Response)**: `resp.GetMemberCashierOrderListResp`, `resp.GetMemberOrderDetailResp`, `resp.GetMemberCashierOrderSearchResp`, `resp.GetMemberOrderManageListResp`, `resp.GetMemberOrderManageDetailResp` 等。

**其他技术栈应用**:

*   **事件总线**: 外送订单状态变更 (接单、拒单、备餐完成) 可能发布事件，通知用户或骑手。
*   **Webhook/消息队列**: 可能与外部配送平台集成，通过 Webhook 或消息队列接收和发送订单状态。

#### 2.11 会员管理与充值订单

**文件**: `main/app/api/v1/cashier/cashier_member.go`

**概述**: 提供收银端会员等级、卡类型查询，会员添加，会员信息充值，以及充值订单的创建、支付、撤销和打印功能。

**控制器层 (`cashier_member.go`) 实现**:

*   **API 接口**:
    *   `/api/v1/cashier/member/level_list`: 获取会员等级列表。
    *   `/api/v1/cashier/member/card_type_list`: 获取会员卡类型列表。
    *   `/api/v1/cashier/member/search_list`: 模糊搜索会员。
    *   `/api/v1/cashier/member/recharge_member_info`: 获取充值会员信息。
    *   `/api/v1/cashier/member/add`: 添加会员。
    *   `/api/v1/cashier/member/recharge_order_doing`: 获取进行中的充值订单。
    *   `/api/v1/cashier/member/recharge_order_create`: 创建充值订单。
    *   `/api/v1/cashier/member/recharge_order_add_payment_method`: 充值订单添加支付方式。
    *   `/api/v1/cashier/member/recharge_order_cancel_payment_method`: 充值订单撤销支付方式。
    *   `/api/v1/cashier/member/recharge_order_confirm`: 确认充值订单。
    *   `/api/v1/cashier/member/recharge_order_print`: 打印充值订单。
    *   `/api/v1/cashier/member/recharge_order_payment_qrcode`: 获取充值订单支付二维码。
    *   `/api/v1/cashier/member/check_password`: 使用会员优惠验证密码。
    *   `/api/v1/cashier/member/handle_member_balance`: 处理会员余额 (仅限内网访问)。
    *   `/api/v1/cashier/member/handle_cash_box_balance`: 处理钱箱余额 (仅限内网访问)。
*   **服务调用**: 调用 `service.IMemberSrv`, `service.IRechargeOrderSrv`, `service.IPaymentMethodSrv`, `service.ICashBoxSrv` 处理会员及充值订单业务逻辑。

**服务层实现**:

*   **`service.IMemberSrv`**: 包含获取会员等级列表、卡类型列表、模糊搜索会员、获取充值会员信息、添加会员、使用会员优惠验证密码，以及处理会员余额等方法。
*   **`service.IRechargeOrderSrv`**: 包含获取进行中的充值订单、创建充值订单、添加/撤销支付方式、确认充值订单、打印充值订单、获取充值订单支付二维码等方法。
*   **`service.IPaymentMethodSrv`**: 用于获取支付方式信息。
*   **`service.ICashBoxSrv`**: 用于处理钱箱余额。

**数据访问层 (Repository) 可能涉及**:

*   **`repository.IMemberRepo`**: 会员数据访问。
*   **`repository.IMemberLevelRepo`**: 会员等级数据访问。
*   **`repository.IMemberCardTypeRepo`**: 会员卡类型数据访问。
*   **`repository.IRechargeOrderRepo`**: 充值订单数据访问。
*   **`repository.IPaymentRepo`**: 充值订单支付单据数据访问。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.AddMemberReq`, `req.RechargeReq`, `req.RechargeOrderAddPaymentMethodReq`, `req.RechargeOrderCancelPaymentMethodReq`, `req.ConfirmRechargeOrder`, `req.PrintRechargeOrderReq`, `req.CheckMemberPasswordReq`, `req.RechargeOrderPaymentQrcodeReq`, `req.MemberBalanceChangeReq`, `req.CashBoxBalanceChangeReq` 等。
*   **响应 (Response)**: `resp.MemberLevelList`, `resp.MemberCardTypeList`, `resp.SearchMemberList`, `resp.RechargeMember`, `resp.RechargeOrder`, `resp.ConfirmRechargeOrder`, `resp.RechargeOrderPaymentQrcodeInfoResp` 等。

**其他技术栈应用**:

*   **内部接口**: `HandleMemberBalance` 和 `HandleCashBoxBalance` 接口可能设计为内部接口，并通过 IP 限制进行安全访问。
*   **事务管理**: 充值订单的创建、支付、余额变动等操作需要严格的事务管理。

#### 2.12 历史订单管理

**文件**: `main/app/api/v1/cashier/cashier_order_old.go`

**概述**: 处理收银端的历史订单查询、取消、删除、打印和退款功能。**该模块通过调用旧的 PHP 后端接口实现大部分功能。**

**控制器层 (`cashier_order_old.go`) 实现**:

*   **API 接口**:
    *   `/api/v1/cashier/order_old/list`: 获取历史订单列表。
    *   `/api/v1/cashier/order_old/info`: 获取历史订单详情。
    *   `/api/v1/cashier/order_old/cancel`: 取消历史订单。
    *   `/api/v1/cashier/order_old/delete`: 删除历史订单。
    *   `/api/v1/cashier/order_old/refund_info`: 获取历史订单退款信息。
    *   `/api/v1/cashier/order_old/refund`: 历史订单退款。
    *   `/api/v1/cashier/order_old/print_receipt`: 打印历史订单小票。
    *   `/api/v1/cashier/order_old/print_invoice`: 打印历史订单发票。
    *   `/api/v1/cashier/order_old/invoice_info`: 获取历史订单发票信息。
*   **服务调用**:
    *   **HTTP 客户端**: 使用 `utils.HttpPost` 或 `utils.HttpGet` 直接调用 `http://nginx/api/cashier/order.order/...` 等 PHP 后端接口。
    *   可能还会调用 `service.IMemberSrv` 处理会员余额，`service.ICashBoxSrv` 处理钱箱余额。

**服务层实现 (在 PHP 后端)**:

*   大部分业务逻辑在 PHP 后端实现。Go 服务层仅作为代理，负责请求转发和结果解析。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.OrderListReq`, `req.OrderInfoReq`, `req.OrderCancelReq`, `req.OrderRefundReq` 等 (这些 DTO 结构体可能需要与 PHP 后端接口的参数保持一致)。
*   **响应 (Response)**: `resp.OrderListPaginationResp`, `resp.OrderInfosResp` 等 (也需与 PHP 后端接口的响应结构一致)。

**其他技术栈应用**:

*   **HTTP 客户端**: 核心是 Go 的 HTTP 客户端调用 PHP 后端接口。
*   **兼容性**: 需处理 Go 与 PHP 后端之间的数据格式转换和错误码映射。

#### 2.13 打印记录管理

**文件**: `main/app/api/v1/cashier/cashier_printer.go`

**概述**: 管理收银端的打印记录，包括查询打印配置、打印列表、以及打印和打印报告功能。

**控制器层 (`cashier_printer.go`) 实现**:

*   **API 接口**:
    *   `/api/v1/cashier/printer/list_condition`: 获取打印列表查询条件。
    *   `/api/v1/cashier/printer/list`: 获取打印列表。
    *   `/api/v1/cashier/printer/print`: 打印。
    *   `/api/v1/cashier/printer/report`: 打印报告。
*   **服务调用**: 调用 `printerService.IPrinterLogSrv` 处理打印记录和打印业务逻辑。

**服务层实现**:

*   **`printerService.IPrinterLogSrv`**: 包含获取打印列表查询条件、获取打印列表、打印、打印报告等方法。

**数据访问层 (Repository) 可能涉及**:

*   **`printerRepository.IPrinterLogRepo`**: 打印日志数据访问。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.PrinterListReq`, `req.PrinterPrintReq`, `req.PrinterReportReqs` 等。
*   **响应 (Response)**: `resp.PrinterLogData`, `resp.PrinterBaseResp`, `resp.PrinterListPaginationResp`, `resp.PrinterData` 等。

**其他技术栈应用**:

*   **打印机集成**: 可能会涉及与各种打印机协议的集成，例如 ESC/POS 命令。

#### 2.14 产品管理

**文件**: `main/app/api/v1/cashier/cashier_product.go`

**概述**: 提供收银端产品和产品类别的查询功能。

**控制器层 (`cashier_product.go`) 实现**:

*   **API 接口**:
    *   `/api/v1/cashier/product/list`: 获取收银产品列表。
    *   `/api/v1/cashier/product/category_list`: 获取收银产品类别列表。
*   **服务调用**: 调用 `service.IProductSrv` 处理产品业务逻辑。

**服务层实现**:

*   **`service.IProductSrv`**: 包含获取收银产品列表、获取收银产品类别列表等方法。

**数据访问层 (Repository) 可能涉及**:

*   **`repository.IProductRepo`**: 产品数据访问。
*   **`repository.IProductCategoryRepo`**: 产品类别数据访问。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.ProductListReq`。
*   **响应 (Response)**: `product_resp.ProductListWithPaginationResp`, `product_resp.ProductCategoryListResp`。

#### 2.15 售罄管理

**文件**: `main/app/api/v1/cashier/cashier_sold_out.go`

**概述**: 管理收银端商品的售罄状态，包括查询售罄商品、设置和取消售罄。

**控制器层 (`cashier_sold_out.go`) 实现**:

*   **API 接口**:
    *   `/api/v1/cashier/sold_out/list`: 获取售罄商品列表。
    *   `/api/v1/cashier/sold_out/set`: 设置商品售罄。
    *   `/api/v1/cashier/sold_out/cancel`: 取消商品售罄。
*   **服务调用**: 调用 `service.ISoldOutSrv` 处理售罄业务逻辑。

**服务层实现**:

*   **`service.ISoldOutSrv`**: 包含获取售罄商品列表、设置商品售罄、取消商品售罄等方法。

**数据访问层 (Repository) 可能涉及**:

*   **`repository.ISoldOutRepo`**: 售罄商品数据访问。
*   **`repository.IProductRepo`**: 可能涉及产品信息。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.SoldOutProductListReq`, `req.SetSoldOutReq`, `req.CancelSoldOutReq`。
*   **响应 (Response)**: `resp.SoldOutProductListResp`。

#### 2.16 统计管理

**文件**: `main/app/api/v1/cashier/cashier_statistics.go`

**概述**: 提供收银端统计报表功能，包括销售统计、支付方式统计等。

**控制器层 (`cashier_statistics.go`) 实现**:

*   **API 接口**:
    *   `/api/v1/cashier/statistics/sales_summary`: 获取销售汇总统计。
    *   `/api/v1/cashier/statistics/sales_detail_list`: 获取销售明细列表。
    *   `/api/v1/cashier/statistics/sales_trend_graph`: 获取销售趋势图数据。
    *   `/api/v1/cashier/statistics/payment_method`: 获取支付方式统计。
*   **服务调用**: 调用 `service.IStatisticsSrv` 处理统计业务逻辑。

**服务层实现**:

*   **`service.IStatisticsSrv`**: 包含获取销售汇总统计、销售明细列表、销售趋势图数据、支付方式统计等方法。

**数据访问层 (Repository) 可能涉及**:

*   **`repository.IOrderRepo`**: 订单数据，用于销售统计。
*   **`repository.IPaymentRepo`**: 支付数据，用于支付方式统计。

**数据传输对象 (DTO) 使用**:

*   **请求 (Request)**: `req.SalesStatisticsReq`, `req.PaymentMethodStatisticsReq`。
*   **响应 (Response)**: `resp.SalesSummaryStatistics`, `resp.SalesDetailList`, `resp.SalesTrendGraph`, `resp.PaymentMethodStatistics`。

**其他技术栈应用**:

*   **数据聚合**: 统计功能通常涉及复杂的数据查询、聚合和计算。
*   **缓存**: 对于访问频率高且实时性要求不极高的统计数据，可以考虑使用缓存来提高性能。

### 3. 公共组件和模式

除了上述特定功能模块的实现外，收银端在实现过程中广泛使用了以下公共组件和设计模式：

*   **数据库管理器 (`pkg/database.DBManager`)**: 统一管理数据库连接，支持多租户数据库连接，通过 `gin.Context` 传递数据库 ID。
*   **缓存 (`pkg/cache.Cache`)**: 用于存储频繁访问的数据，例如系统配置、商品信息等，提高读取性能。
*   **并发锁 (`pkg/lock.Lock`)**: 在处理敏感操作 (如库存扣减、余额变动、桌台状态变更) 时，使用 UUID 锁或其他并发锁机制保证数据一致性。
*   **事件总线 (`pkg/eventbus`)**: 用于解耦模块，例如订单状态变更事件、打印任务事件等，可以实现异步处理和跨模块通信。
*   **错误处理 (`ttpos-server-go/app/errors`)**: 统一的错误包装机制，方便错误追踪和日志记录。
*   **认证鉴权 (`middleware.Auth`)**: 基于 JWT Token 的认证中间件，保护 API 接口。
*   **国际化 (`ttpos-server-go/i18n`)**: 提供多语言支持，使系统能够适应不同地区的用户。
*   **日志 (`go.uber.org/zap`)**: 结构化日志记录，便于故障排查和系统监控。
*   **统一响应格式**: 所有 API 接口返回 `dto.Response` 结构，包含 `code`, `message`, `data` 字段。
*   **依赖注入**: 服务层通过构造函数接收依赖，提高代码可测试性和可维护性。
*   **Repository 模式**: 将数据访问逻辑从服务层中抽象出来，方便数据库切换和单元测试。

### 4. 总结

收银端功能模块的实现设计遵循了 Go 语言的开发规范和最佳实践，利用 Gin 框架构建 RESTful API，并通过服务层、数据访问层、DTO 等分层架构实现了业务逻辑的清晰分离。同时，充分利用了公共组件和设计模式来提高系统的健壮性、可扩展性和可维护性。特别是在订单处理、结账流程、并发控制和与外部系统的集成方面，都进行了细致的设计和实现考虑。
