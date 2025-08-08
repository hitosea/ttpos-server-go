import { $get, $post } from '../request';

export interface authorizationListType {
  keyword: string;
  page: number;
  list_rows: number;
  configured?: number;
}

export interface authorizationListTypeItem {
  uuid: number;
  name: string;
  link_phone: string;
  admin: string;
  erpnext_site_code: string;
  erpnext_company_abbr: string;
  erpnext_branch_name: string;
}

export interface erpnextSiteCodeItem {
  name: string;
  code: string;
}

// 商家授权列表
export function getAuthorizationList(params: authorizationListType) {
  return $get('/erpnext/index', { params });
}

// 获取ERPNext站点编码列表
export function getErpnextSiteCode() {
  return $get('/erpnext/siteCode');
}

export interface erpnextSiteCompanyParams {
  site_code: string;
  company_abbr?: string;
}

export interface erpnextSiteCompanyItem {
  company_name: string;
  company_abbr: string;
}

// 获取ERPNext站点公司名称
export function getErpnextSiteCompany(params: erpnextSiteCompanyParams) {
  return $get('/erpnext/siteCompany', { params });
}

export interface erpnextAddParams {
  uuid: string;
  erpnext_site_code: string;
  erpnext_company_abbr: string;
}

export function erpnextAdd(data: erpnextAddParams) {
  return $post('/erpnext/add', data);
}
