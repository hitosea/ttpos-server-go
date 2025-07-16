import { $get } from '../request';

export type searchParams = {
  uuid: number;
  channel: string;
  month: string;
  page: number;
  list_rows: number;
};

export function getTakeoutOrder(params: searchParams) {
  return $get('delivery/ledger', { params });
}
