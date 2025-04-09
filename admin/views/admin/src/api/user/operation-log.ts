import { HttpResponsePage, $post } from '../request';

// 列表
export interface OperationLogType {
  username?: string; // 用户名
  page?: number; // 查询页数
  list_rows?: number; // 查询条数
}
export interface OperationLogData {
  id?: number;
  admin_user_id?: number; // 用户id
  title?: string; // 标题
  url?: string; // 访问url
  request_type?: string; // 请求类型
  browser?: string; // 浏览器
  agent?: string; // 浏览器信息
  content?: string; // 操作内容
  ip?: string; // 登录ip
  create_time?: number; // 创建时间
  username?: string; // 用户名
  real_name?: string; // 姓名
}
export function getOperationLog(data: OperationLogType) {
  return $post<{
    list?: HttpResponsePage<OperationLogData[]>;
  }>('/admin.optlog/index', data);
}
