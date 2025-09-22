import request from '@/utils/request';

let MarketingApi = {
    /*活动列表*/
    activityList(data, errorback) {
        return request._get('/shop/marketing.activity/list', data, errorback);
    },
    /*添加活动*/
    activityAddGet(data, errorback) {
        return request._get('/shop/marketing.activity/add', data, errorback);
    },
    /*添加活动*/
    activityAdd(data, errorback) {
        return request._post('/shop/marketing.activity/add', data, errorback);
    },
    /*编辑活动*/
    activityEditGet(data, errorback) {
        return request._get('/shop/marketing.activity/edit', data, errorback);
    },
    activityEdit(data, errorback) {
        return request._post('/shop/marketing.activity/edit', data, errorback);
    },

    /*失效活动*/
    activityDisable(data, errorback) {
        return request._post('/shop/marketing.activity/disable', data, errorback);
    },

    /*活动记录*/
    activityRecord(data, errorback) {
        return request._get('/shop/marketing.activity/record', data, errorback);
    },

    /*优惠券列表*/
    couponList(data, errorback) {
        return request._get('/shop/marketing.coupon/list', data, errorback);
    },

    /*添加优惠券*/
    couponAddGet(data, errorback) {
        return request._post('/shop/marketing.coupon/add', data, errorback);
    },

    /*编辑优惠券*/
    couponEdit(data, errorback) {
        return request._post('/shop/marketing.coupon/edit', data, errorback);
    },

    /*优惠券记录*/
    couponRecord(data, errorback) {
        return request._get('/shop/marketing.coupon/record', data, errorback);
    },

    /*活动优惠券列表*/
    activityCouponList(data, errorback) {
        return request._get('/shop/marketing.activity/couponList', data, errorback);
    },

    /*修改优惠券状态*/
    couponStatus(data, errorback) {
        return request._post('/shop/marketing.coupon/status', data, errorback);
    },

    /*删除优惠券*/
    couponDelete(data, errorback) {
        return request._post('/shop/marketing.coupon/delete', data, errorback);
    },

};

export default MarketingApi;