import { $get } from '../request';

export interface logListType {
  /**
   * 商户 UUID(平台管理员可指定，商户管理员只能查自己的)
   */
  company_uuid?: string;
  /**
   * 导入类型(1-TTPOS推送到平台 2-平台推送到TTPOS)，0 查询所有
   */
  import_type?: number;
  /**
   * 页码，默认 1
   */
  page?: number;
  /**
   * 每页数量，默认 20，最大 100
   */
  page_size?: number;
  /**
   * 外卖平台(grab/lineman等)，为空查询所有
   */
  platform?: string;
  /**
   * 导入状态(0-进行中 1-成功 2-失败)，-1 查询所有
   */
  status?: number;
}

export interface LogListItem {
  uuid?: number;
  platform?: string;
  import_type?: number;
  import_direction?: string;
  status?: number;
  progress?: number;
  success_count?: number;
  failure_count?: number;
  total_count?: number;
  error_message?: string;
  start_time?: number;
  end_time?: number;
  duration?: number;
  createtime?: number;
}

export function getLogList(params: logListType) {
  return $get('/takeout/logs', { params });
}
