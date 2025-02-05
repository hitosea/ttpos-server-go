import { UserInfoType } from '@/api/login';
import { getCache, setCache, removeCache } from './cache';

export function isLogin(): boolean {
  return !!getToken();
}

export function getToken() {
  return getCache('token');
}

export function setToken(data: string) {
  return setCache('token', data);
}

export function clearAuthInfo() {
  removeCache('token');
  removeCache('user_info');
}

export function getUserInfo() {
  return getCache('user_info');
}

export function setUserInfo(data: UserInfoType) {
  return setCache('user_info', data);
}
