import { HttpResponsePage, $post, $get } from '../request';

// 列表
export interface RoleListType {
  keyword?: string; // 角色名称
  page?: number; // 查询页数
  list_rows?: number; // 查询条数
}
export interface RoleListData {
  id?: number;
  role_name?: string; // 角色名称
  sort?: number; // 排序(数字越小越靠前)
  create_time?: string; // 创建时间
  update_time?: string; // 更新时间
}
export function getRoleList(data: RoleListType) {
  return $post<{
    list?: HttpResponsePage<RoleListData[]>;
  }>('/admin.role/index', data);
}

// 删除
export function fetchDelete(id?: number) {
  return $post('/admin.role/delete', { id });
}

// 添加
export interface RoleEditType {
  id?: number;
  role_name?: string; // 角色名称
  access_id?: number[]; // 权限id数组
  sort?: number;
}
export function fetchRoleAdd(data: RoleEditType) {
  return $post('/admin.role/add', data);
}
// 获取信息
export function getRoleAdd() {
  return $get('/admin.role/add');
}

// 编辑
export function fetchRoleEdit(data: RoleEditType) {
  return $post('/admin.role/edit', data);
}
// 获取信息
export function getRoleEdit(id?: number) {
  return $get(`/admin.role/edit?id=${id}`);
}
