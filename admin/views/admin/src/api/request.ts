import axios from 'axios';
import type { AxiosRequestConfig, InternalAxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';

import { closeGlobalLoading, globalLoading, message } from '@/utils/feedback';
import { getToken, clearAuthInfo } from '@/utils/auth';
import { router } from '@/router/index';
import { $t, getLanguage } from '@/i18n';

import { secretKeyStore } from '@/stores/secretKey';
import { encrypt, decrypt } from '@/utils/jsencrypt';

// 配置扩展
export interface HttpRequestConfig {
  isMessageError?: boolean; // 是否开启错误提示弹窗
  isLoading?: boolean; // 是否开启加载
  isDownload?: boolean; // 是否为下载
  encrypt?: boolean; // 是否加密
}

// 返回数据结构
export interface HttpResponse<T = any> {
  code?: number;
  data?: T;
  msg?: string;
}

// 分页数据结构
export interface HttpResponsePage<T = any> {
  data?: T; // 数据
  total?: number; // 总数
  current_page?: number; // 当前页
  per_page?: number; // 每页条数
}

/**
 * // 实例化
 * const controller = new AbortController();
 * // 请求
 * axios.ge$t('url', { signal: controller.signal }).then(function(response) {});
 * // 取消请求
 * controller.abort()
 */

// 创建一个axios实例
const request = axios.create({
  baseURL: '/api/admin/',
  timeout: 60 * 1000, // 请求超时时间 默认60秒
  headers: {
    'Content-Type': 'application/json;charset=UTF-8',
  },
});

/**
 * @desc: 异常拦截处理器
 * @param { Object } error 错误信息
 */
const handlerError = (error?: AxiosError, options?: HttpRequestConfig & AxiosRequestConfig): Promise<AxiosError> => {
  // 登录失效，删除token，跳转到登录页
  if ((error as any)?.code == -2 || (error as any)?.code == -1) {
    clearAuthInfo();
    router.replace('/login');
  }
  if (options?.isMessageError || error?.code) {
    if (error?.code === 'ERR_CANCELED') {
      message.error($t('请求被已被手动取消'));
    } else if (error?.code === 'ECONNABORTED') {
      message.error($t('请求超时，请重试'));
    } else if (error?.code === 'ERR_NETWORK') {
      message.error($t('网络异常，请检查您的网络设置'));
    } else {
      if ((error as any)?.msg || (error as any)?.message) message.error((error as any).msg || (error as any)?.message);
    }
  }
  return Promise.reject(error);
};

/**
 * @desc: 请求发送前拦截
 * @param { Object } config 配置参数
 */
request.interceptors.request.use(async (options: HttpRequestConfig & InternalAxiosRequestConfig) => {
  // 设置默认值
  const config: HttpRequestConfig & InternalAxiosRequestConfig = {
    isMessageError: true, // 是否开启错误提示弹窗
    isLoading: false, // 是否开启加载
    isDownload: false, // 是否为下载
    encrypt: false, // 是否加密
    ...options,
  };
  // 设置token
  if (getToken()) config.headers['token'] = getToken();
  // 设置语言
  if (getLanguage()) config.headers['language'] = getLanguage();
  //浏览器缓存获取UUID+时间戳
  if (localStorage.getItem('uuid')) {
    config.headers['Suuid'] = localStorage.getItem('uuid') + '--' + new Date().getTime();
  }
  // 默认开启loading
  if (config.isLoading) globalLoading();
  // 捕获400、500
  config.validateStatus = (status) => {
    if (config?.isMessageError && status >= 400) message.error(`E${status} - ${$t('服务器繁忙，请稍候再试')}`);
    //
    return status >= 200 && status < 500;
  };
  // 是否为下载
  if (config.isDownload) {
    config.responseType = 'arraybuffer';
  }
  // 是否加密
  if (config.encrypt && config.data && ['post', 'put'].includes(config.method || '')) {
    const serverSecretKey = await secretKeyStore().getServerSecretKey();
    if (serverSecretKey.type === 'jsencrypt' && serverSecretKey.key) {
      // 加密数据
      const data: { encrypted?: string } = {};
      const encryptData = await encrypt(config.data, serverSecretKey.key);
      if (encryptData) data.encrypted = encryptData;
      config.data = data;
      // 生成秘钥对，给后台传公钥
      const clientSecretKey = await secretKeyStore().getClientSecretKey();
      config.headers.encrypt = `encrypt_type=${serverSecretKey.type};encrypt_id=${serverSecretKey.id};client_type=${clientSecretKey.type};client_key=${clientSecretKey.publicKey}`;
    }
  }

  return config;
}, handlerError);

/**
 * @desc: 服务端响应后拦截
 * @param { Object } response 返回的数据
 */
request.interceptors.response.use(async (response: AxiosResponse) => {
  // 所有参数
  const config: HttpRequestConfig & InternalAxiosRequestConfig = response.config;
  // 关闭loading
  if (config.isLoading) closeGlobalLoading();
  // 数据处理
  let data = response.data;
  // 是否为下载
  if (config.isDownload) {
    try {
      data = JSON.parse(new TextDecoder('utf-8').decode(new Uint8Array(data)));
    } catch (err) {
      // 获取类型类型和编码
      const blob = new Blob([data], {
        type: response.headers['content-type'],
      });
      // 获取文件名
      const fileArr = response.headers['content-disposition'].spli$t('=');
      const fileName = decodeURI(fileArr[fileArr.length - 1].replace(/^(utf-8'')/, '')).replace(/"/g, '');
      const link = document.createElement('a');
      link.download = fileName;
      link.style.display = 'none';
      link.href = URL.createObjectURL(blob);
      document.body.appendChild(link);
      link.click();
      URL.revokeObjectURL(link.href);
      document.body.removeChild(link);
    }
    return data;
  }
  // 解密
  if (config.encrypt && ['post', 'put'].includes(config.method || '') && data.encrypted) {
    const clientSecretKey = await secretKeyStore().getClientSecretKey();
    data = await decrypt(data.encrypted, clientSecretKey.privateKey);
  }
  // 返回成功
  if (data.code == 1) return data;
  // 提示错误
  return handlerError(data, config);
}, handlerError);

// $get
export const $get = <T = any, D = any>(url: string, config?: HttpRequestConfig & AxiosRequestConfig<D>): Promise<HttpResponse<T>> => request.get(url, config);

// $post
export const $post = <T = any, D = any>(url: string, data?: D, config?: HttpRequestConfig & AxiosRequestConfig<D>): Promise<HttpResponse<T>> => request.post(url, data, config);

// $put
export const $put = <T = any, D = any>(url: string, data?: D, config?: HttpRequestConfig & AxiosRequestConfig<D>): Promise<HttpResponse<T>> => request.put(url, data, config);

// $delete
export const $delete = <T = any, D = any>(url: string, config?: HttpRequestConfig & AxiosRequestConfig<D>): Promise<HttpResponse<T>> => request.delete(url, config);

// 默认导出
export default request;
