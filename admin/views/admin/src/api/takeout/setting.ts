import { $post, $get } from '../request';

export interface DistanceRange {
  end: number;
  price_per_km: number;
  is_unlimited: boolean;
}

export interface TakeoutSettingData {
  channel?: string;
  basic_fee?: number;
  base_delivery_fee?: number;
  rider_acceptance_timeout?: number;
  distance_range?: DistanceRange[];
  uuid?: string; // 添加可选的 uuid 字段
  channels?: string; // 添加可选的 channels 字段
}

export function getTakeoutSetting() {
  return $get<TakeoutSettingData[]>('delivery/index');
}

export function fetchTakeoutSetting(data: { channels: TakeoutSettingData[] }) {
  return $post('delivery/index', data);
}
