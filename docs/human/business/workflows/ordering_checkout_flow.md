# 完整的点餐下单送厨出菜结账流程设计文档

#### 1. 简介

本文档旨在详细梳理本项目中完整的点餐、下单、送厨、出菜和结账的核心业务流程。该流程涵盖了用户从浏览商品到最终完成支付的整个生命周期，涉及多个模块和组件的紧密协作。通过事件总线、并发锁和分层架构等设计，确保了流程的高效、可靠和可扩展性。

#### 2. 主要模块和组件

以下是此核心业务流程中涉及的主要模块和组件：

*   **前端/API 网关**: 负责用户界面交互，接收用户的点餐、支付等请求，并将请求路由到后端服务。
*   **控制器层 (`main/app/api/v1/cashier`)**: 接收 API 请求，进行参数绑定和初步验证，然后调用相应的服务层方法处理业务逻辑。
*   **服务层 (`main/app/service`)**: 封装核心业务逻辑，协调多个数据存储库（Repository），执行复杂的操作。
*   **数据传输对象 (`main/app/dto`)**: 定义请求 (`req`) 和响应 (`resp`) 的数据结构，用于在各层之间传递数据。
*   **数据访问层 (`main/app/repository`)**: 封装对数据库的 CRUD 操作，提供统一的数据访问接口。
*   **数据模型 (`main/app/model`)**: 定义数据库表对应的 Go 结构体。
*   **数据库管理器 (`main/pkg/database`)**: 管理数据库连接的生命周期、连接池和多租户数据库的动态获取。
*   **事件总线 (`main/pkg/eventbus`)**: 实现模块间的解耦，通过发布和订阅事件来触发异步操作和跨模块通信。
*   **并发锁 (`main/pkg/lock`)**: 用于在高并发场景下保证数据一致性，例如在操作库存或桌台状态时。
*   **日志 (`main/pkg/logger`)**: 记录系统运行状态、业务操作和错误信息，便于监控和故障排查。
*   **配置 (`main/config`)**: 存储系统各项配置参数，如数据库连接、API 密钥等。
*   **中间件 (`main/middleware`)**: 提供通用的请求处理逻辑，如认证 (`Auth`)、权限验证、异常捕获等。

#### 3. 完整业务流程

##### 3.1 点餐流程 (Ordering Process)

点餐流程是用户与系统互动的第一步，主要包括商品浏览、选择、数量调整、规格属性选择、备注等操作。

1.  **商品查询**:
    *   **前端**: 用户在收银端或 H5 端浏览商品列表。
    *   **API**: `GET /cashier/product/list` (获取产品列表) 和 `GET /cashier/product/category/list` (获取产品类别列表)。
    *   **控制器 (`ProductHandler`)**: 接收请求，绑定 `req.ProductListReq` 参数，调用 `productSrv.GetProductList` 或 `productSrv.GetProductCategoryList`。
    *   **服务 (`ProductSrv`)**: 查询产品和产品类别信息，可能涉及缓存 (`pkg/cache`)，然后返回给控制器。
    *   **数据库 (`ProductRepository`, `ProductCategoryRepository`)**: 从数据库中获取商品和分类数据。

2.  **选择商品并加入购物车**:
    *   **前端**: 用户选择商品，调整数量，选择规格属性（例如大小、甜度等），添加商品备注。
    *   **API**: 
        *   `POST /cashier/desk/order/cart/product/add` (向购物车添加商品)
        *   `POST /cashier/desk/order/cart/product_package/add` (向购物车添加套餐)
        *   `GET /cashier/desk/order/cart/product/flavor_and_attribute` (查询购物车商品规格/属性)
        *   `POST /cashier/desk/order/cart/product/flavor_and_attribute` (修改购物车商品规格/属性)
        *   `POST /cashier/desk/order/cart/product/num` (修改购物车商品数量)
        *   `POST /cashier/desk/order/product/remark` (桌台订单商品备注)
        *   `POST /cashier/desk/order/remark` (整单备注)
    *   **控制器 (`DeskHandler`)**: 接收请求，绑定相应的 `req` 参数，调用 `orderSrv` 中对应的方法。
    *   **服务 (`OrderSrv`)**:
        *   处理商品/套餐的添加、修改、删除逻辑。
        *   进行库存的初步校验。
        *   **并发锁 (`pkg/lock`)**: 在更新购物车或订单相关数据时，可能会使用并发锁（例如基于桌台 UUID 的锁）来防止并发冲突，确保数据一致性。
        *   维护当前订单的购物车状态（可能存储在内存或 Redis 中）。
        *   更新销售账单（如果当前没有销售账单，则创建一个新的）。
    *   **数据模型 (`model.SaleBill`, `model.SaleOrder`, `model.SaleOrderItem`)**: 更新销售账单、订单和订单商品项的数据。

3.  **开台 (如果桌台模式)**:
    *   **前端**: 如果是桌台模式，用户在点餐前需要选择桌台或开台。
    *   **API**: `POST /cashier/desk/open` (开台)。
    *   **控制器 (`DeskHandler`)**: 接收 `req.DeskOrderCreateReq` 参数，调用 `deskSrv.CreateDeskOrder`。
    *   **服务 (`DeskSrv`)**:
        *   更新桌台状态。
        *   创建新的销售账单 (`model.SaleBill`)。
        *   发布 `EventOpenDesk` 事件。
    *   **事件总线**: 发布 `EventOpenDesk` 事件，通知其他模块（如实时更新桌台状态的前端）。
    *   **并发锁 (`pkg/lock`)**: 在开台或修改桌台状态时，使用桌台 UUID 锁来防止并发操作。

##### 3.2 下单流程 (Placing Order Process)

下单流程是将用户购物车中的商品正式转化为订单，并准备进行支付。

1.  **订单检查**:
    *   **前端**: 用户点击“结账”或“确认订单”按钮前。
    *   **API**: `GET /cashier/desk/order/check` (订单检查)。
    *   **控制器 (`DeskHandler`)**: 接收 `req.InstantOrderCheckReq` 参数，调用 `orderSrv.OrderCheck`。
    *   **服务 (`OrderSrv`)**:
        *   进行最终的商品库存检查。
        *   检查必点商品是否选择完成。
        *   检查商品价格是否发生变化。
        *   可能返回 `OrderCheckRes` 包含检查结果（例如库存不足的商品列表）。
    *   **错误处理**: 如果检查不通过，返回相应的错误码和提示信息，可能包含需要用户确认的数据。

2.  **创建销售订单**:
    *   **前端**: 订单检查通过后，用户确认下单。
    *   **API**: `POST /cashier/desk/order/sale_order/create` (创建一个销售订单)。
    *   **控制器 (`DeskHandler`)**: 接收 `req.InstantOrderSaleOrderCreateReq` 参数，调用 `orderSrv.InstantOrderSaleOrderCreate`。
    *   **服务 (`OrderSrv`)**:
        *   根据购物车信息，创建或更新 `model.SaleOrder` 和 `model.SaleOrderItem` 记录。
        *   扣除商品库存。
        *   发布 `EventCreateMemberSaleOrder` 等相关事件（如果是会员外送订单）。
        *   **并发锁 (`pkg/lock`)**: 在扣减库存时使用并发锁，确保库存的准确性。
    *   **事件总线**: 发布订单创建相关的事件。

##### 3.3 送厨流程 (Sending to Kitchen Process)

送厨流程是将已下单的商品信息发送到厨房，通知厨师开始制作。

1.  **商品送厨**:
    *   **前端**: 用户点击“送厨”按钮。可能支持分批送厨。
    *   **API**: 
        *   `POST /cashier/desk/order/cart/cooking` (送厨购物车商品)
        *   `GET /cashier/desk/order/cart/batch/cooking` (获取分批送厨弹框的销售订单商品列表)
        *   `POST /cashier/desk/order/cart/batch/cooking` (分批送厨)
    *   **控制器 (`DeskHandler`)**: 接收 `req.OrderCartProductCookingReq` 或 `req.OrderCartProductBatchCookingReq` 参数，调用 `orderSrv.InstantOrderCartProductCooking` 或 `orderSrv.OrderCartProductBatchCooking`。
    *   **服务 (`OrderSrv`)**:
        *   更新订单商品的制作状态。
        *   触发厨房打印。
        *   发布 `EventSentCooking` (送厨事件) 或 `EventSentCookingPre` (预送厨事件)。
    *   **事件总线**: 发布送厨相关的事件，厨房事件处理器会订阅这些事件，进行打印和状态更新。
    *   **打印服务**: 打印厨房订单（可能通过集成第三方打印机或内部打印服务）。

##### 3.4 出菜流程 (Serving Process)

出菜流程是厨房完成商品制作后，通知服务员上菜。

1.  **完成制作**:
    *   **厨房**: 厨师完成菜品制作。
    *   **API (可能由厨房端调用)**: 内部 API 或通过其他机制触发。
    *   **事件总线**: 发布 `EventFinishMenu` (完成制作事件)。
    *   **服务 (`OrderSrv`)**:
        *   订阅 `EventFinishMenu` 事件。
        *   更新订单商品的制作状态为“已完成”。
        *   通知服务员上菜（可能通过 WebSocket 或其他即时通信方式）。

##### 3.5 结账流程 (Checkout Process)

结账流程是订单的最后阶段，涉及支付、优惠、抹零、会员折扣和最终的订单完结。

1.  **获取结账信息**:
    *   **前端**: 用户进入结账页面。
    *   **API**: `GET /cashier/desk/order/payment/info` (获取结账页面信息)。
    *   **控制器 (`DeskHandler`)**: 接收 `req.InstantOrderPaymentInfoReq` 参数，调用 `orderSrv.InstantOrderPaymentInfo`。
    *   **服务 (`OrderSrv`)**:
        *   计算订单最终金额，包括商品价格、服务费、税费等。
        *   查询可用的优惠券、积分、会员折扣等信息。
        *   返回包含所有支付信息的 `resp.InstantOrderPaymentInfoResp`。

2.  **应用优惠/积分/会员**:
    *   **前端**: 用户选择使用优惠券、抵扣积分，或享受会员折扣。
    *   **API**: 
        *   `POST /cashier/desk/order/payment/coupon` (选择或取消优惠券)
        *   `POST /cashier/desk/order/payment/points` (设置订单的抵扣积分数量)
        *   `GET /cashier/desk/order/member/discount` (获取订单会员优惠)
        *   `POST /cashier/desk/order/member/confirm` (确认使用会员优惠并验证密码)
        *   `DELETE /cashier/desk/order/member/cancel` (不使用此会员)
        *   `POST /cashier/desk/order/discount` (桌台订单打折，包含改价、打折、抹零)
        *   `POST /cashier/desk/order/discount/cancel` (取消桌台订单所有优惠折扣)
        *   `POST /cashier/desk/order/payment/zero_rule` (设置结账抹零规则)
        *   `POST /cashier/desk/order/free` (免单)
    *   **控制器 (`DeskHandler`)**: 接收相应的 `req` 参数，调用 `orderSrv` 中对应的方法。
    *   **服务 (`OrderSrv`, `MemberSrv`)**:
        *   根据用户选择的优惠、积分、会员等级等，重新计算订单金额。
        *   进行会员密码验证。
        *   更新订单的优惠信息。
        *   处理改价、打折、抹零、免单等操作。
        *   **并发锁 (`pkg/lock`)**: 在涉及到金额或会员信息修改时，使用并发锁保证数据一致性。
        *   发布 `EventDiscountSaleOrder`, `EventDiscountZeroSaleOrder`, `EventDiscountChangePriceSaleOrder`, `EventCancelSaleOrderDiscount`, `EventFreeSaleOrder` 等事件。
    *   **事件总线**: 发布折扣优惠相关的事件。

3.  **创建支付单并完成支付**:
    *   **前端**: 用户选择支付方式，确认支付。
    *   **API**: 
        *   `GET /cashier/desk/order/payment/qrcode` (获取支付方式的二维码信息)
        *   `POST /cashier/desk/order/payment/create` (创建一个支付单)
        *   `POST /cashier/desk/order/payment/cancel` (撤销一个支付单)
        *   `POST /cashier/desk/order/payment/finish` (完成销售订单的付款结账)
    *   **控制器 (`DeskHandler`)**: 接收 `req.InstantOrderPaymentCreateReq` 或 `req.InstantOrderPaymentFinishReq` 等参数，调用 `orderSrv` 中对应的方法。
    *   **服务 (`OrderSrv`)**:
        *   生成支付订单，与第三方支付平台交互（例如生成支付二维码）。
        *   处理支付结果回调。
        *   **并发锁 (`pkg/lock`)**: 在支付过程中使用锁来避免重复支付或并发问题。
        *   更新订单支付状态。
        *   发布 `EventCheckoutSaleOrder` (结账事件) 和 `EventPayFinishMemberSaleOrder` (支付完成事件)。
    *   **事件总线**: 发布支付和结账相关的事件，触发后续的统计、库存扣减、会员积分变动等异步操作。

4.  **订单完结与后续处理**:
    *   **服务 (`OrderSrv`)**:
        *   订阅 `EventCheckoutSaleOrder` 和 `EventPayFinishMemberSaleOrder` 事件。
        *   更新订单最终状态为“已完成”。
        *   扣减最终库存。
        *   触发销售统计 (`EventStatisticsSale`)。
        *   触发会员积分变动 (`EventChangeMemberPoints`)。
        *   打印小票或发票 (`POST /cashier/desk/order/print`, `POST /cashier/desk/order/print/invoice`)。
        *   如果存在订单反结账 (`EventOrderReverseSettle`)、取消订单 (`EventCancelOrder`) 等操作，也会通过事件总线进行处理。

#### 4. 总结

完整的点餐下单送厨出菜结账流程是一个复杂且关键的业务链条，本项目通过以下设计原则和技术手段，确保了其高效、稳定运行：

*   **清晰的分层架构**: 控制器、服务、仓库分离，职责明确。
*   **事件驱动**: 利用事件总线实现模块间松耦合，支持异步处理和灵活扩展。
*   **并发控制**: 在关键业务操作中引入并发锁，保证数据一致性。
*   **GORM ORM**: 简化数据库操作，提高开发效率。
*   **完善的错误处理与日志**: 确保业务流程中的错误能够被妥善处理和记录，便于故障排查。
*   **多租户支持**: 数据库管理器能够管理多个商家数据库，适应多租户业务模式。

通过以上设计，系统能够高效地处理从用户点餐到最终结账的全部流程，为餐饮业务提供稳定可靠的后端支持。
