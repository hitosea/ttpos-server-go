import { HttpResponsePage, $post, $get } from '../request';

// 商家列表
export interface ShopListType {
  keyword?: string; // 商家名称/ID
  shop_type?: 0 | 1 | 2; // 商家类型 0全部, 1连锁店, 2单店
  status?: 0 | 1 | 2; // 状态 0全部, 1启用, 2禁用
  page?: number; // 查询页数
  list_rows?: number; // 查询条数
}

export interface ShopListData {
  app_id?: number; // 小程序id
  expire_time?: string; // 过期时间
  auth_day?: number; // 授权时间(天) 0为永不过期
  status?: 0 | 1; // 状态 1=启用 0=禁用
  create_time?: string; // 创建时间
  id?: number; // 列表ID
  suppliers_total?: number; // 子店铺总数, 商家类型通过这个来判断显示
  suppliers?: ShopListData[]; // 子列表
  user_name?: string; // 超管用户名
  shop_supplier_id?: number; // 线上店铺id
  shop_supplier_name?: string; // 线上店铺名称
  offline_shop_supplier_id?: number; // 线下店铺id
  parent_id?: number; // 父id
  link_phone?: number; // 联系电话
  logo?: string; // logo
  level?: number; // 层级
  sale_stock?: 0 | 1; // 进销存 0=不开启 1=开启
  reserve?: 0 | 1; // 预订 0=不开启 1=开启
  cash_limit?: number; // 收银机上限
  kitchen_limit?: number; // 厨显上限
  tablet_limit?: number; // 平板上限
  address?: string; // 联系地址
  store_type?: 10 | 20; // 店铺类型10加盟20自营
  category_set?: 10 | 20; // 商品分类设置10同步 主店20分店创建
}

// 商家列表
export function getShopList(data: ShopListType) {
  return $post<{
    list?: HttpResponsePage<ShopListData[]>;
  }>('/shop/index', data);
}

// 删除商家
export function fetchShopDelete(app_id?: number) {
  return $post('/shop/delete', { app_id });
}

// 启用禁用状态
export function fetchUpdateStatus(app_id?: number) {
  return $post('/shop/updateStatus', { app_id });
}

// 启用禁用进销存
export function fetchUpdateSaleStock(app_id?: number) {
  return $post('/shop/updateSaleStock', { app_id });
}

// 启用禁用预定
export function fetchUpdateReserve(app_id?: number) {
  return $post('/shop/updateReserve', { app_id });
}

// 获取授权码
export function getLicense(app_id?: number) {
  return $post('/shop/getLicense', { app_id });
}

// 更新授权码
export function fetchUpdateLicense(app_id?: number) {
  return $post('/shop/updateLicense', { app_id });
}

export interface ShowAddType {
  name?: string; // 商家名称
  level?: number; // 商家等级: 从1开始
  parent_id?: number; // 上级商家ID
  store_type?: number; // 经营模式: 20直营店, 10加盟店
  category_set?: number; // 商品分类设置10同步 主店20分店创建
  logo?: string; // Logo
  sale_stock?: number; // 进销存: 0不开启, 1开启
  is_open_tax?: number; // 开启税务
  reserve?: number; // 预订: 0不开启, 1开启
  auth_day?: number; // 授权时间(天) 0为永不过期
  auth_start_time?: string; //
  link_phone?: string; // 联系电话
  cash_limit?: number; // 收银机上限
  kitchen_limit?: number; // 厨显上限
  tablet_limit?: number; // 平板上限
  assistant_limit?: number; // 平板上限
  user_name?: string; // 超管用户名
  password?: string; // 超管密码
  confirm_password?: string; // 再次确认
  address?: string; // 所在地
  mac_addr?: string; //mac地址
  serial_number?: string; //服务序列号
  deploy_mode?: number; // 部署方式
  chain_number?: string; // 连锁编号
  suppliers_total?: number; // 连锁编号
  languages?: any; // 支持语言
  timezone?: string; // 时区
  table_limit?: number; // 桌台限制
  printer_limit?: number; // 打印机限制
  is_open_member?: number; // 是否开启会员: 0不开启, 1开启
  is_open_tablet?: number; // 是否开启平板: 0不开启, 1开启
  is_open_scan?: number; // 是否开启扫码H5: 0不开启, 1开启
  is_open_assistant?: number; // 是否开启点餐助手: 0不开启, 1开启
  is_open_kitchen_kds?: number; // 是否开启后厨KDS: 0不开启, 1开启
  is_open_buffet?: number; // 是否开启自助餐: 0不开启, 1开启
  is_accept_scan_order?: number; // 是否开启扫码点餐接单: 0不开启, 1开启
  is_open_local_print?: number; // 是否开启本地打印服务: 0不开启, 1开启（v1.1.0）
  enable_sms?: number; // 是否开启短信服务: 0不开启, 1开启
  sms_quota?: number; // 短信额度
  is_open_coupon?: number; // 是否开启优惠券: 0不开启, 1开启
  is_open_marketing?: number; // 是否开启营销活动: 0不开启, 1开启
  is_open_advanced_ticket_print?: number; // 是否开启高级票据打印: 0不开启, 1开启
  coordinates?: string; // 经纬度
  enable_table_map?: number; // 是否启用桌台地图能力: 0不开启, 1开启
  enable_grab_delivery?: number; // 是否开启Grab外卖: 0不开启, 1开启
  enable_lineman_delivery?: number; // 是否开启LINE MAN外卖: 0不开启, 1开启
  enable_data_management?: number; // 是否启用数据管理能力: 0不开启, 1开启
  enable_kiosk?: number; // 是否开启自助点餐机: 0不开启, 1开启
  is_open_member_instant?: number; // 是否开启扫码点餐到店自取: 0不开启, 1开启
}

// 添加商家
export function fetchShowAdd(data: ShowAddType) {
  return $post('/shop/add', data);
}

// 编辑商家
export function fetchShowEdit(data: ShowAddType | { app_id?: number }) {
  return $post('/shop/edit', data);
}

// 商家详情
export function getShowEdit(app_id?: number) {
  return $get(`/shop/edit?app_id=${app_id}`);
}

// 商家下拉列表
export interface SelectListType {
  keyword?: string; // 商家名称/ID
  page?: number; // 查询页数
  list_rows?: number; // 查询条数
}
export interface SelectListData {
  id?: number; // id
  name?: string; // 商家名称
}

export function getSelectList(data: SelectListType) {
  return $post<{
    list?: HttpResponsePage<SelectListData[]>;
  }>('/shop/selectList', data);
}

export interface LangListType {
  keyword?: string; //
  page?: number; // 查询页数
  list_rows?: number; // 查询条数
}

export function getLangList(data: LangListType) {
  return $post<any>('/shop/langData', data);
}
// 获取信息
export function getPayment(shop_supplier_id?: number) {
  return $get(`/shop/payment?shop_supplier_id=${shop_supplier_id}`);
}

export function setPayment(data: any) {
  return $post<any>('/shop/payment', data);
}
