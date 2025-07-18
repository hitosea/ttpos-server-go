import request from '@/utils/request';
let OrderApi = {
  /*订单列表*/
  takeOrderlist(data, errorback) {
    return request._post('/shop/takeout.order/index', data, errorback);
  },
  storeOrderlist(data, errorback) {
    return request._post('/shop/store.order/index', data, errorback);
  },
  /*订单详情*/
  takeOrderdetail(data, errorback) {
    return request._post('/shop/takeout.order/detail', data, errorback);
  },
  storeOrderdetail(data, errorback) {
    return request._post('/shop/store.order/detail', data, errorback);
  },
  /*取消*/
  storeConfirm(data, errorback) {
    return request._post('/shop/store.Operate/orderCancel', data, errorback);
  },
  /*删除*/
  storedelete(data, errorback) {
    return request._post('/shop/store.operate/delete', data, errorback);
  },
  takeConfirm(data, errorback) {
    return request._post('/shop/takeout.Operate/orderCancel', data, errorback);
  },
  /*退款*/
  takeRefund(data, errorback) {
    return request._post('/shop/takeout.Operate/refund', data, errorback);
  },

  storeRefund(data, errorback) {
    return request._post('/shop/store.operate/orderRefund', data, errorback);
  },
  getStoreRefund(data, errorback) {
    return request._get('/shop/store.operate/orderRefund', data, errorback);
  },
  orderRefundAgain(data, errorback) {
    return request._get('/shop/store.operate/orderRefundAgain', data, errorback);
  },
  /*确认收货并退款*/
  takeReceipt(data, errorback) {
    return request._post('/shop/takeout.refund/receipt', data, errorback);
  },
  storeReceipt(data, errorback) {
    return request._post('/shop/store.refund/receipt', data, errorback);
  },

  orderProductList(data, errorback) {
    return request._post('/shop/store.operate/orderProductList', data, errorback);
  },
  /*核销*/
  takeExtract(data, errorback) {
    return request._post('/shop/takeout.operate/extract', data, errorback);
  },
  storeExtract(data, errorback) {
    return request._post('/shop/store.operate/extract', data, errorback);
  },
  sendDada(data, errorback) {
    return request._post('/shop/takeout.operate/sendOrder', data, errorback);
  },
  deliveryData(data, errorback) {
    return request._get('/shop/takeout.delivery/index', data, errorback);
  },
  storeExport(data, errorback) {
    return request._get('/shop/store.operate/export', data, errorback);
  },

  // 充值订单
  getRechargeOrder(data, errorback) {
    return request._post('/shop/store.UserRechargeOrder/index', data, errorback);
  },
  getRechargeOrderDetail(data, errorback) {
    return request._post('/shop/store.UserRechargeOrder/detail', data, errorback);
  },
  getRechargeOrderCancel(data, errorback) {
    return request._post('/shop/store.UserRechargeOrder/cancel', data, errorback);
  },
  getRechargeOrderRefund(data, errorback) {
    return request._get('/shop/store.UserRechargeOrder/refund', data, errorback);
  },
  postRechargeOrderRefund(data, errorback) {
    return request._post('/shop/store.UserRechargeOrder/refund', data, errorback);
  },
  //充值订单重新退款
  postRechargeOrderRefundAgain(data, errorback) {
    return request._post('/shop/store.UserRechargeOrder/refundAgain', data, errorback);
  },

  //外卖订单列表
  postTakeoutOrderList(data, errorback) {
    return request._post('/shop/store.MemberOrder/index', data, errorback);
  },

  //外卖订单详情
  postTakeoutOrderDetail(data, errorback) {
    return request._post('/shop/store.MemberOrder/detail', data, errorback);
  },
};

export default OrderApi;
