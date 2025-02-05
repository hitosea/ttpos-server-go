import { $post, $get } from '../request';

// 登录
export interface UserInfoType {
  app_id?: number;
  currency?: { unit: string; is_open: 1 | 0; vices: { vice_unit: string; unit_rate: number } };
  is_open?: 1 | 0;
  unit?: string;
  vices?: { vice_unit: string; unit_rate: number };
  unit_rate?: number;
  vice_unit?: string;
  logoUrl?: string;
  shop_name?: string;
  shop_supplier_id?: number;
  supplier_name?: string;
  token?: string;
  user_name?: string;
  user_type?: 0;
  version?: number;
}
export interface LoginDataType {
  username?: string; // 账号
  password?: string; // 密码
  code?: string; // 图形验证码
}

export function login(data: LoginDataType) {
  return $post<UserInfoType>('/passport/login', data, { encrypt: true });
}

// 退出登录
export function logout() {
  return $post('/passport/logout');
}

// 修改密码
export interface EditPasswordType {
  oldPass?: string; // 旧密码
  pass?: string; // 新密码
  checkPass?: string; // 确认新密码
}
export function editPassword(data: EditPasswordType) {
  return $post('/passport/renew', data);
}

// 获取验证码
export function getCaptcha(data: any) {
  return $post('/passport/captcha', data);
}

// 登录页logo
export function getLoginLogo() {
  return $get('passport/logo');
}

// 客户端KEY
export function getPublicKey(data: any) {
  return $post('passport/getKey', data);
}

// 基础信息
export function getService() {
  return $post('setting.service/info');
}
