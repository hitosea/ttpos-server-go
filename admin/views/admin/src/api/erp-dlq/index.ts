import { $get, $post } from '../request';

export interface FailedSIStats {
  save_failed_count: number;
  cancel_failed_count: number;
  return_failed_count: number;
  total_failed_count: number;
  save_pending_count: number;
  cancel_pending_count: number;
  return_pending_count: number;
  total_pending_count: number;
}

export interface SIRetryParams {
  site_code?: string;
  record_id?: number;
  sale_order_uuid?: string;
  msg_type?: string;
  action?: 'retry' | 'skip';
}

export interface SIRetryResult {
  retry_count: number;
}

// 获取 SI 失败统计
export function getSIFailedStats(params?: { site_code?: string }) {
  return $get<FailedSIStats>('/erpnext/siFailedStats', { params });
}

// 重试失败的 SI
export function retrySI(data: SIRetryParams) {
  return $post<SIRetryResult>('/erpnext/siRetry', data);
}
