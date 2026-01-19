import { HttpResponsePage, $post } from '../request';

// 列表
export interface StaffListType {
  keyword?: string; // 搜索关键词（邮箱、手机号、员工ID）
  page?: number; // 查询页数
  list_rows?: number; // 查询条数
}

// 门店角色信息
export interface CompanyRoleInfo {
  company_uuid: number; // 门店UUID
  company_name: string; // 门店名称
  roles: Array<{
    role_uuid: number; // 角色UUID
    role_name: string; // 角色名称
  }>;
}

export interface StaffListData {
  uuid?: number; // 员工UUID
  email?: string; // 邮箱
  phone?: string; // 手机号
  real_name?: string; // 姓名
  is_disable?: 0 | 1; // 是否禁用（0-未禁用，1-禁用）
  last_company_uuid?: number; // 上次登录新管理端的商家UUID
  create_time?: string; // 创建时间
  update_time?: string; // 更新时间
  company_list?: CompanyRoleInfo[]; // 关联的门店列表和角色信息
}

export function getStaffList(data: StaffListType) {
  return $post<{
    list?: HttpResponsePage<StaffListData[]>;
  }>('/admin.staff/index', data);
}

// 启用禁用状态
export function fetchUpdateStaffStatus(uuid?: number) {
  return $post('/admin.staff/updateStatus', { uuid });
}

// 添加
export interface StaffAddType {
  email?: string; // 邮箱（全平台唯一）
  phone?: string; // 手机号（全平台唯一，允许空字符串）
  real_name?: string; // 姓名
  password?: string; // 登录密码
  confirm_password?: string; // 确认密码
  company_uuid?: number; // 关联的门店UUID
  role_uuids?: number[]; // 在该门店的角色UUID列表
}
export function fetchStaffAdd(data: StaffAddType) {
  return $post('/admin.staff/add', data);
}

// 编辑
export interface StaffEditType extends StaffAddType {
  uuid?: number; // 员工UUID
  company_list?: Array<{
    company_uuid: number | undefined; // 门店UUID
    role_uuids?: number[]; // 在该门店的角色UUID列表
  }>; // 关联的门店列表（可选，用于编辑时更新门店关联）
}
export function fetchStaffEdit(data: StaffEditType) {
  return $post('/admin.staff/edit', data);
}
