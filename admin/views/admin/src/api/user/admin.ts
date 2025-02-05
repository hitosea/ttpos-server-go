import { HttpResponsePage, $post } from '../request';
import { type RoleListData } from './role';

// 列表
export interface AdminListType {
  keyword?: string; // 用户名,姓名, 用户ID
  page?: number; // 查询页数
  list_rows?: number; // 查询条数
}
export interface AdminUserRole {
  admin_user_id: number;
  create_time: string;
  id: number;
  role: RoleListData;
  role_id: number;
  update_time: string;
}
export interface AdminListData {
  admin_user_id?: number; // 主键id
  user_name?: string; // 用户名
  real_name?: string; // 姓名
  is_super?: number; // 是否超级管理员
  status?: 0 | 1; // 状态(0未启用,1已启用)
  create_time?: string; // 创建时间
  update_time?: string; // 更新时间
  delete_time?: string; // 删除时间
  userRole?: AdminUserRole[]; // 角色列表
}
export function getAdminList(data: AdminListType) {
  return $post<{
    list?: HttpResponsePage<AdminListData[]>;
  }>('/admin.user/index', data);
}

// 启用禁用状态
export function fetchUpdateStatus(admin_user_id?: number) {
  return $post('/admin.user/updateStatus', { admin_user_id });
}

// 删除
export function fetchDelete(admin_user_id?: number) {
  return $post('/admin.user/delete', { admin_user_id });
}

// 添加
export interface AdminAddType {
  user_name?: string; // 用户名
  phone?: string; // 手机号
  password?: string; // 登录密码
  confirm_password?: string; // 确认密码
  real_name?: string; // 姓名
  role_id?: number[]; // 角色ids
}
export function fetchAdminAdd(data: AdminAddType) {
  return $post('/admin.user/add', data);
}

// 编辑
export function fetchAdminEdit(data: AdminAddType | { admin_user_id?: number }) {
  return $post('/admin.user/edit', data);
}
