import { HttpResponsePage, $post } from '../request';

// 列表
export interface LoginLogType {
  username?: string; // 用户名
  page?: number; // 查询页数
  list_rows?: number; // 查询条数
}
export interface LoginLogData {
  id?: number;
  admin_user_id?: number; // 用户id
  username?: string; // 用户名
  ip?: string; // 登录ip
  result?: string; // 登录结果
  create_time?: number; // 创建时间
}

export function getLoginLog(data: LoginLogType) {
  return $post<{
    list?: HttpResponsePage<LoginLogData[]>;
  }>('/admin.loginlog/index', data);
}
