# ttpos 与 LINE MAN 集成用户故事/用例

本文档概述了 ttpos 系统与 LINE MAN 平台进行集成时的关键用户故事和用例，涵盖了认证、菜单同步、订单处理和门店状态控制等方面。

## 1. 认证相关用例

### 用例名称：ttpos 获取 LINE MAN 访问令牌

**角色：** ttpos 系统

**目标：** 成功获取用于访问 LINE MAN API 的访问令牌。

**前置条件：**

*   ttpos 已配置有效的 Client ID 和 Client Secret。
*   LINE MAN 认证服务正常运行。

**主流程：**

1.  ttpos 向 LINE MAN 认证接口发送 OAuth 2.0 Client Credentials 请求，包含 Client ID 和 Client Secret。
2.  LINE MAN 认证服务验证凭据。
3.  LINE MAN 认证服务返回一个访问令牌（Access Token）。
4.  ttpos 接收并存储访问令牌，用于后续 API 调用。

**备选流程：**

*   **备选 1.1：凭据无效**
    1.  LINE MAN 认证服务验证凭据失败。
    2.  LINE MAN 认证服务返回错误响应（例如，401 Unauthorized）。
    3.  ttpos 记录错误并触发警报，可能尝试重新配置或通知管理员。

*   **备选 1.2：服务不可用**
    1.  ttpos 无法连接到 LINE MAN 认证服务。
    2.  ttpos 记录错误并触发警报，可能进行重试或通知管理员。

**后置条件：**

*   **成功：** ttpos 拥有有效的 LINE MAN 访问令牌。
*   **失败：** ttpos 未能获取访问令牌，并且已记录错误或触发警报。

---

### 用例名称：LINE MAN 验证 ttpos API 请求

**角色：** LINE MAN 系统

**目标：** 成功验证 ttpos 发送的 API 请求的合法性。

**前置条件：**

*   ttpos 已获取有效的访问令牌。
*   LINE MAN 接收到来自 ttpos 的 API 请求，请求头中包含访问令牌。
*   LINE MAN 认证服务正常运行。

**主流程：**

1.  LINE MAN 接收到 ttpos 的 API 请求。
2.  LINE MAN 从请求头中提取访问令牌。
3.  LINE MAN 认证服务验证访问令牌的有效性（例如，是否过期，是否被撤销）。
4.  如果令牌有效，LINE MAN 授权 ttpos 执行请求的操作。

**备选流程：**

*   **备选 2.1：访问令牌无效或缺失**
    1.  LINE MAN 验证访问令牌失败（例如，令牌过期、无效格式或缺失）。
    2.  LINE MAN 返回错误响应（例如，401 Unauthorized）。
    3.  ttpos 接收到错误，并可能尝试重新获取访问令牌。

*   **备选 2.2：IP 白名单限制（如果启用）**
    1.  LINE MAN 检查 ttpos 的请求 IP 地址是否在白名单中。
    2.  如果 IP 不在白名单中，LINE MAN 拒绝请求并返回错误响应。
    3.  ttpos 接收到错误并触发警报。

**后置条件：**

*   **成功：** LINE MAN 成功验证 ttpos 请求，并允许处理后续业务逻辑。
*   **失败：** LINE MAN 拒绝 ttpos 请求，并返回相应的错误信息。

## 2. 菜单同步相关用例

### 用例名称：ttpos 推送完整菜单快照至 LINE MAN

**角色：** ttpos 系统（商家操作员）

**目标：** 将 ttpos 中的最新完整菜单数据（包括菜品、选项、价格、状态等）成功同步到 LINE MAN 平台，确保 LINE MAN 上展示的菜单与 ttpos 保持一致。

**前置条件：**

*   ttpos 已完成对菜单的修改（例如，添加新菜品、修改价格、更新库存状态等）。
*   ttpos 已获取有效的 LINE MAN 访问令牌。
*   LINE MAN 菜单同步 API 可用。

**主流程：**

1.  ttpos 内部检测到菜单数据发生变化，或商家操作员在 ttpos 后台手动触发菜单同步。
2.  ttpos 构建包含所有菜单详情的完整菜单快照数据。
3.  ttpos 调用 LINE MAN 的 "Menu sync" API，并将菜单快照数据作为请求体发送。
4.  LINE MAN 接收到菜单同步请求并开始处理。
5.  LINE MAN 处理完成后，（可选）通过 "Menu sync notification" API 向 ttpos 发送同步结果通知（成功或失败）。
6.  ttpos 接收到同步结果通知，并根据结果更新内部同步状态。

**备选流程：**

*   **备选 1.1：LINE MAN 菜单同步失败**
    1.  LINE MAN 处理菜单同步请求时发生错误。
    2.  LINE MAN 通过 "Menu sync notification" API 向 ttpos 发送失败通知，包含错误详情。
    3.  ttpos 记录失败日志，并触发警报通知商家操作员，可能提供重试选项。

*   **备选 1.2：网络或 API 调用失败**
    1.  ttpos 调用 "Menu sync" API 时遇到网络问题或 LINE MAN API 服务不可用。
    2.  ttpos 记录错误并触发警报，可能进行重试或通知管理员。

**后置条件：**

*   **成功：** LINE MAN 上的菜单已更新为 ttpos 的最新完整菜单数据。ttpos 内部同步状态更新为成功。
*   **失败：** LINE MAN 上的菜单未能成功更新。ttpos 内部同步状态标记为失败，并已记录错误信息。

---

### 用例名称：LINE MAN 请求 ttpos 重新同步菜单

**角色：** LINE MAN 系统

**目标：** 在 LINE MAN 检测到菜单数据可能不一致时，触发 ttpos 重新推送完整菜单，以修复潜在的数据差异。

**前置条件：**

*   LINE MAN 检测到其内部菜单数据可能与 ttpos 不一致，或需要强制刷新。
*   ttpos 已注册并监听 LINE MAN 的 "Trigger sync menu" API。
*   ttpos 已获取有效的 LINE MAN 访问令牌。

**主流程：**

1.  LINE MAN 调用 ttpos 的 "Trigger sync menu" API，请求 ttpos 重新同步菜单。
2.  ttpos 接收到 "Trigger sync menu" 请求。
3.  ttpos 立即启动菜单同步流程，构建并调用 LINE MAN 的 "Menu sync" API，推送完整的菜单快照。
4.  LINE MAN 接收并处理 ttpos 推送的菜单快照。
5.  LINE MAN 处理完成后，（可选）通过 "Menu sync notification" API 向 ttpos 发送同步结果通知（成功或失败）。
6.  ttpos 接收到同步结果通知，并根据结果更新内部同步状态。

**备选流程：**

*   **备选 2.1：ttpos 无法响应触发同步请求**
    1.  LINE MAN 调用 "Trigger sync menu" API 时，ttpos 服务不可用或响应超时。
    2.  LINE MAN 记录错误，可能进行重试或触发内部警报。

*   **备选 2.2：ttpos 重新同步失败**
    1.  ttpos 在收到 "Trigger sync menu" 后执行菜单同步时发生错误。
    2.  （可选）LINE MAN 收到 "Menu sync notification" 失败通知。
    3.  LINE MAN 可能再次触发 "Trigger sync menu" 或通过其他方式处理。

**后置条件：**

*   **成功：** ttpos 已成功重新推送菜单快照，LINE MAN 上的菜单数据与 ttpos 保持一致。
*   **失败：** ttpos 未能成功重新同步菜单，LINE MAN 上的菜单数据可能仍存在不一致。

---

### 用例名称：ttpos 更新单个菜单项状态至 LINE MAN（可选）

**角色：** ttpos 系统（商家操作员）

**目标：** 商家操作员在 ttpos 中更新单个菜品（或选项）的可售状态时，将此变更实时同步至 LINE MAN 平台。

**前置条件：**

*   商家操作员在 ttpos 后台修改了某个菜单项（或选项）的状态（例如，售罄、恢复可售、暂停销售）。
*   ttpos 已获取有效的 LINE MAN 访问令牌。
*   LINE MAN 更新菜单项状态 API 可用。

**主流程：**

1.  商家操作员在 ttpos 后台将某个菜品（或选项）标记为 "售罄" (SOLD_OUT_TODAY)，"恢复可售" (AVAILABLE) 或 "暂停销售" (SUSPENDED)。
2.  ttpos 构建包含菜单项 UUID 和新状态的请求数据。
3.  ttpos 调用 LINE MAN 的 "Update menu item status" (或 "Update menu propertyValue status") API，并发送状态更新请求。
4.  LINE MAN 接收并处理状态更新请求。
5.  LINE MAN 成功更新对应菜单项的状态。

**备选流程：**

*   **备选 3.1：LINE MAN API 调用失败**
    1.  ttpos 调用 LINE MAN 状态更新 API 时遇到网络问题或 LINE MAN API 错误。
    2.  ttpos 记录错误并触发警报，可能进行重试或通知商家操作员。

*   **备选 3.2：菜单项在 LINE MAN 不存在**
    1.  LINE MAN 接收到状态更新请求，但无法找到对应的菜单项。
    2.  LINE MAN 返回错误响应。
    3.  ttpos 记录错误并触发警报。

**后置条件：**

*   **成功：** LINE MAN 上对应菜单项（或选项）的状态已更新为 ttpos 中的最新状态。
*   **失败：** LINE MAN 上对应菜单项（或选项）的状态未能成功更新，ttpos 已记录错误信息。

## 3. 订单处理相关用例

### 用例名称：LINE MAN 推送新订单详情至 ttpos

**角色：** LINE MAN 系统

**目标：** 将顾客在 LINE MAN 平台下单的详细信息实时推送给 ttpos 系统，以便 ttpos 进行订单处理和后续的厨房出餐流程。

**前置条件：**

*   顾客已在 LINE MAN 平台成功提交订单。
*   ttpos 已注册并监听 LINE MAN 的 "Place order notification" API。
*   ttpos 服务正常运行。

**主流程：**

1.  顾客在 LINE MAN 平台完成下单。
2.  LINE MAN 系统调用 ttpos 的 "Place order notification" API，将新订单的详细信息（例如，订单 ID、商品列表、价格、配送信息、支付状态等）作为请求体发送。
3.  ttpos 接收到新订单通知。
4.  ttpos 验证订单数据的完整性和合法性。
5.  ttpos 将订单数据保存到数据库，并更新内部订单状态。
6.  ttpos 通知厨房或备餐系统准备订单。

**备选流程：**

*   **备选 1.1：ttpos 接收订单失败**
    1.  LINE MAN 推送订单时，ttpos 服务不可用、响应超时或返回处理失败。
    2.  LINE MAN 记录错误，可能进行重试或触发内部警报。
    3.  （可选）LINE MAN 可能通知商家订单未能成功推送到 POS。

*   **备选 1.2：订单数据格式错误或不完整**
    1.  ttpos 接收到订单通知，但订单数据格式不符合预期或缺少关键信息。
    2.  ttpos 返回错误响应给 LINE MAN，并记录错误日志。
    3.  LINE MAN 收到错误后，可能进行数据修正或触发警报。

**后置条件：**

*   **成功：** ttpos 系统已成功接收、保存并开始处理 LINE MAN 推送的新订单。
*   **失败：** ttpos 未能成功处理新订单通知，LINE MAN 可能需要重试或人工介入。

---

### 用例名称：LINE MAN 推送订单状态更新通知至 ttpos（可选）

**角色：** LINE MAN 系统

**目标：** 在 LINE MAN 平台订单状态发生变化时（例如，订单完成、订单取消），实时通知 ttpos 系统，以保持订单状态同步。

**前置条件：**

*   LINE MAN 平台上的订单状态发生变化（例如，骑手已取餐、订单已送达、顾客取消订单、商家取消订单）。
*   ttpos 已注册并监听 LINE MAN 的 "Status update notification" API。
*   ttpos 服务正常运行。

**主流程：**

1.  LINE MAN 平台上的某个订单状态更新。
2.  LINE MAN 调用 ttpos 的 "Status update notification" API，将订单 ID 和最新的状态（例如，已完成、已取消）作为请求体发送。
3.  ttpos 接收到订单状态更新通知。
4.  ttpos 验证订单 ID 和状态的合法性。
5.  ttpos 更新其内部数据库中对应订单的状态。
6.  ttpos 可能触发相应的业务逻辑（例如，释放库存、更新财务记录）。

**备选流程：**

*   **备选 2.1：ttpos 接收状态更新失败**
    1.  LINE MAN 推送状态更新时，ttpos 服务不可用、响应超时或返回处理失败。
    2.  LINE MAN 记录错误，可能进行重试或触发内部警报。

*   **备选 2.2：订单在 ttpos 不存在或状态不匹配**
    1.  ttpos 接收到状态更新通知，但无法找到对应的订单 ID，或收到的状态与当前内部状态不一致。
    2.  ttpos 返回错误响应，并记录错误日志。
    3.  LINE MAN 收到错误后，可能触发警报或人工介入进行核对。

**后置条件：**

*   **成功：** ttpos 系统中对应订单的状态已更新为 LINE MAN 的最新状态。
*   **失败：** ttpos 未能成功更新订单状态，订单状态可能不一致。

---

### 用例名称：LINE MAN 推送订单编辑更新通知至 ttpos（可选）

**角色：** LINE MAN 系统

**目标：** 在顾客或客服在 LINE MAN 平台编辑订单（例如，修改菜品数量、移除菜品、修改配送地址等）后，实时通知 ttpos 系统，以确保订单信息的同步。

**前置条件：**

*   LINE MAN 平台上的订单被编辑并成功保存。
*   ttpos 已注册并监听 LINE MAN 的 "Order update notification" API。
*   ttpos 服务正常运行。

**主流程：**

1.  LINE MAN 平台上的某个订单被编辑。
2.  LINE MAN 调用 ttpos 的 "Order update notification" API，将更新后的完整订单详情或变更部分作为请求体发送。
3.  ttpos 接收到订单更新通知。
4.  ttpos 验证订单数据的完整性和合法性。
5.  ttpos 更新其内部数据库中对应订单的信息。
6.  ttpos 可能触发重新计算订单总价、重新通知厨房等业务逻辑。

**备选流程：**

*   **备选 3.1：ttpos 接收订单更新失败**
    1.  LINE MAN 推送订单更新时，ttpos 服务不可用、响应超时或返回处理失败。
    2.  LINE MAN 记录错误，可能进行重试或触发内部警报。

*   **备选 3.2：订单在 ttpos 不存在或数据冲突**
    1.  ttpos 接收到订单更新通知，但无法找到对应的订单 ID，或更新数据与当前内部数据存在冲突。
    2.  ttpos 返回错误响应，并记录错误日志。
    3.  LINE MAN 收到错误后，可能触发警报或人工介入进行核对。

**后置条件：**

*   **成功：** ttpos 系统中对应订单的详细信息已更新为 LINE MAN 的最新状态。
*   **失败：** ttpos 未能成功更新订单信息，订单数据可能不一致。

## 4. 门店状态控制相关用例

### 用例名称：ttpos 控制 LINE MAN 门店营业状态（可选）

**角色：** ttpos 系统（商家操作员）

**目标：** 商家操作员在 ttpos 系统中设置门店的营业状态（开店/关店）时，将此操作同步到 LINE MAN 平台，从而控制门店在 LINE MAN 上的可见性和接单能力。

**前置条件：**

*   商家操作员在 ttpos 后台决定修改门店的营业状态。
*   ttpos 已获取有效的 LINE MAN 访问令牌。
*   LINE MAN 强制开店/关店 API 可用。

**主流程：**

1.  商家操作员在 ttpos 后台将门店设置为 "营业中" (开店) 或 "休息中" (关店)。
2.  ttpos 构建包含门店 ID 和目标营业状态的请求数据。
3.  ttpos 调用 LINE MAN 的 "Force close/open restaurant" API，并发送状态变更请求。
4.  LINE MAN 接收并处理门店营业状态更新请求。
5.  LINE MAN 更新其平台上的门店营业状态，使其对顾客可见或不可见，并相应地启用或禁用接单功能。

**备选流程：**

*   **备选 1.1：LINE MAN API 调用失败**
    1.  ttpos 调用 LINE MAN 营业状态更新 API 时遇到网络问题或 LINE MAN API 错误。
    2.  ttpos 记录错误并触发警报，可能进行重试或通知商家操作员。

*   **备选 1.2：门店在 LINE MAN 不存在或状态冲突**
    1.  LINE MAN 接收到状态更新请求，但无法找到对应的门店 ID，或收到的状态与当前内部状态存在冲突。
    2.  LINE MAN 返回错误响应。
    3.  ttpos 记录错误并触发警报。

**后置条件：**

*   **成功：** LINE MAN 平台上对应门店的营业状态已更新为 ttpos 中的最新状态。
*   **失败：** LINE MAN 平台上对应门店的营业状态未能成功更新，ttpos 已记录错误信息。
