import { $get, HttpResponsePage, $post } from '../request';

export interface TakeoutShopData {
  name: string;
  value: string;
}

export interface TakeoutShopListData {
  channel: string;
  keyword: string;
  page: number;
  list_rows: number;
}

export interface TakeoutShopListDataItem {
  uuid: number;
  name: string;
  delivery_status: number;
  delivery_config: [
    {
      channel: string;
      basic_fee: number;
      base_delivery_fee: number;
      rider_acceptance_timeout: number;
      distance_range: [
        {
          end: number;
          price_per_km: number;
          is_unlimited: boolean;
        },
        {
          end: number;
          price_per_km: number;
          is_unlimited: boolean;
        },
      ];
      config_type: string;
    },
  ];
  channel_names: string[];
  channel_config_types: string[];
}

export interface TakeoutShopSelectData {
  uuid: number;
  name: string;
  channel_names: string[];
}

export interface TakeoutShopAddEditData {
  uuid: number;
  channels: TakeoutShopAddEditDataItem[];
}

export interface TakeoutShopAddEditDataItem {
  channel: string;
  config_type: string;
  basic_fee: number;
  base_delivery_fee: number;
  rider_acceptance_timeout: number;
  distance_range: Array<{
    end: number;
    price_per_km: number;
  }>;
}

export interface CanTakeoutShopListData {
  list_rows?: number;
  keyword?: string;
  configured?: number;
}

// 获取渠道
export function getChannels(params: { configured?: number }) {
  return $get<TakeoutShopData[]>('delivery/channels', { params });
}

// 获取列表
export function getCompanyList(params: TakeoutShopListData) {
  return $get<{
    list?: HttpResponsePage<TakeoutShopListDataItem[]>;
  }>('delivery/companyList', { params });
}

// 更新状态
export function updateStatus(uuid: number) {
  return $post('delivery/updateStatus', { uuid });
}

// 获取可选商家列表
export function getCompanySelect(params: CanTakeoutShopListData) {
  return $get<{ list?: HttpResponsePage<TakeoutShopSelectData[]> }>('delivery/companySelect', { params });
}

// 添加\修改商家外送配置
export function addEditCompany(data: TakeoutShopAddEditData) {
  return $post('delivery/company', data);
}
