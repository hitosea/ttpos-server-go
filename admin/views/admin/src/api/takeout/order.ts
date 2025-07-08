import { $get } from '../request';

export type searchParams = {
  uuid: number;
  channel: string;
  month: string;
};

export function getTakeoutOrder(params: searchParams) {
  return $get('delivery/ledger', { params });
}
