import { $post } from '../request';

// 客户端列表搜索
export interface cashierListType {
  type?: string; //类型：1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端
  version_number?: string; // 版本号
  forced_update?: '' | '1' | '2'; // 状态 0全部, 1是, 2否
  page?: number; // 查询页数
  list_rows?: number; // 查询条数
}

// 客户端列表
export interface dataList {
  type?: string; // 类型：1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端
  brand?: string; //品牌
  version_number?: string; // 版本号
  forced_update?: string; // 强制更新 0否 1是
  package_url?: string; // 包地址
  update_log?: string; // 更新日志
  uuid?: string; // uuid，用于下载
  download_num?: string; // 下载数量
  id?: string; // id
}

// 添加客户端
export interface addDataType {
  type?: string; // 类型：1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端
  brand?: number; //品牌
  version_name?: string; // 版本号
  version_number?: string; // 版本号
  forced_update?: number; // 强制更新 0否 1是
  package_url?: string; // 包地址
  update_log?: string; // 更新日志
}

export interface detailData {
  type?: string; // 类型：1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端
  brand?: number; //品牌
  version_name?: string; // 版本名称
  version_number?: string; // 版本号
  forced_update?: number; // 强制更新 0否 1是
  package_url?: string; // 包地址
  update_log?: string; // 更新日志
  id?: string; // id
  is_publish?: string;
}

// 商家列表
export function getClientList(data: cashierListType) {
  return $post('/client.client/index', data);
}

// 添加
export function addClient(data: addDataType) {
  return $post('/client.client/add', data);
}
// 添加
export function publishClient(data: addDataType) {
  return $post('/client.client/publish', data);
}

// 删除
export function deleteClient(data: dataList) {
  return $post('/client.client/delete', data);
}
// 二维码
export function qrcodeClient(data: dataList) {
  return $post('/client.client/qrcode', data);
}
// 二维码
export function aiTranslate(data: dataList) {
  return $post('/client.client/aiTranslate', data);
}
