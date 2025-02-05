import request from '@/utils/request';

let StoreApi = {
    /*桌位列表*/
    tablelist(data, errorback) {
        return request._post('/shop/store.table.table/index', data, errorback);
    },
    /*桌位添加*/
    addTable(data, errorback) {
        return request._post('/shop/store.table.table/add', data, errorback);
    },
    /*编辑桌位*/
    editTable(data, errorback) {
        return request._post('/shop/store.table.table/edit', data, errorback);
    },
    /*删除桌位*/
    deleteTable(data, errorback) {
        return request._post('/shop/store.table.table/delete', data, errorback);
    },
    /*解除桌位*/
    unbindTable(data, errorback) {
        return request._post('/shop/store.table.table/unbind', data, errorback);
    },
    /*区域列表*/
    arealist(data, errorback) {
        return request._post('/shop/store.table.area/index', data, errorback);
    },
    /*区域添加*/
    addArea(data, errorback) {
        return request._post('/shop/store.table.area/add', data, errorback);
    },
    /*编辑区域*/
    editArea(data, errorback) {
        return request._post('/shop/store.table.area/edit', data, errorback);
    },
    /*删除区域*/
    deleteArea(data, errorback) {
        return request._post('/shop/store.table.area/delete', data, errorback);
    },
    /*类型列表*/
    typelist(data, errorback) {
        return request._post('/shop/store.table.type/index', data, errorback);
    },
    /*类型添加*/
    addType(data, errorback) {
        return request._post('/shop/store.table.type/add', data, errorback);
    },
    /*编辑类型*/
    editType(data, errorback) {
        return request._post('/shop/store.table.type/edit', data, errorback);
    },
    /*删除类型*/
    deleteType(data, errorback) {
        return request._post('/shop/store.table.type/delete', data, errorback);
    },
    /*店内概况*/
    storeSurvey(data, errorback) {
        return request._post('/shop/store.survey/index', data, errorback);
    },

    /*二维码*/
    getQRcode(data, errorback) {
        return request._post('/shop/store.table.table/qrcode', data, errorback);
    },

    /*批量二维码*/
    batchQRcode(data, errorback) {
        return request._post('/shop/store.table.table/batchQrcode', data, errorback);
    },

    /*桌台开关状态*/
    switchStatus(data, errorback) {
        return request._post('/shop/store.table.table/switchStatus', data, errorback);
    },
    /*桌台批量删除*/
    batchTableDelete(data, errorback) {
        return request._post('/shop/store.table.table/batchDelete', data, errorback);
    },
};

export default StoreApi;
