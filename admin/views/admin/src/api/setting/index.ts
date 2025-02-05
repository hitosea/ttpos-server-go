import { $post, $get } from '../request';

// 系统设置
export interface SettingServiceType {
  brand_name?: string; // 品牌名称
  brand_logo?: string; // 品牌LOGO（正方型）
  brand_logo_long?: string; // 品牌LOGO（长方型）
  expiration_reminder?: number; // 到期提醒 0为永久
  member_default_avatar?: string; // 会员默认头像
  browser_logo?: string; // 会员默认头像
  browser_title?: string; // 会员默认头像
  auth_code_bind_validity_period?: number; // 会员默认头像
}

export interface versionType {
  client_version?: string; // 客户端号
  server_version?: string; // 服务器号
}

export function fetchSettingService(data: SettingServiceType) {
  return $post('/setting.service/index', data);
}

export function getSettingService() {
  return $get<{
    vars?: {
      values?: SettingServiceType;
      version?: versionType;
    };
  }>('/setting.service/index');
}

// 获取菜单权限
export interface RoleListData {
  api_path?: string;
  children?: RoleListData[];
  create_time?: string;
  icon?: string;
  id?: number;
  is_menu?: number;
  is_route?: number;
  is_show?: number;
  is_supplier?: number;
  name?: string;
  parent_id?: number;
  path?: string;
  redirect_name?: string;
  remark?: string;
  sort?: number;
  update_time?: string;
}
export function getUserRoleList() {
  return $get<{ menus?: RoleListData[] }>('access/getRoleList');
}
