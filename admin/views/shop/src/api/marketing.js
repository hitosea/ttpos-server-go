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

};

export default MarketingApi;