import request from '@/utils/request';

let IndexApi = {
  /*基础配置*/
  base(data, errorback) {
    return request._post('/shop/index/base', data, errorback);
  },
  /*基础配置*/
  public(data, errorback) {
    return request._post('/shop/index/public', data, errorback);
  },

  /*商城首页*/
  getCount(data, errorback) {
    return request._post('/shop/Index/index', data, errorback);
  },

  /*设备解绑*/
  unbind(data, errorback) {
    return request._post('/shop/index/unbind', data, errorback);
  },

  /*设备解绑*/
  getBindList(data, errorback) {
    return request._post('/shop/index/bindList', data, errorback);
  },

  /*授权码验证*/
  verifyAuthCode(data, errorback) {
    return request._post('/shop/index/verifyAuthCode', data, errorback);
  },

  /*SAAS首次登录修改密码*/
  saasEditPassword(data, errorback) {
    return request._post('/shop/passport/saasEditPassword', data, errorback);
  },

  /*翻译*/
  aiTranslate(data, errorback) {
    return request._post('/shop/index/aiTranslate', data, errorback);
  },

  /*获取验证码*/
  getCaptcha(data, errorback) {
    return request._get('/shop/passport/captcha', data, errorback);
  },

  /*检查名称唯一性*/
  checkNameExist(data, errorback) {
    return request._post('/shop/index/checkNameExist', data, errorback);
  },

  /*获取商品列表*/
  getProductList(data, errorback) {
    return request._post('/shop/index/productList', data, errorback);
  },

  
};

export default IndexApi;
