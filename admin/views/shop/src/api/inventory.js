import request from '@/utils/request'

let InventoryApi = {

    /*库存管理*/
    getErpInventory(data, errorback) {
        return request._post('/shop/inventory.ErpInventory/list', data, errorback);
    },
    
    /*出库记录*/
    getErpInventoryRecordOut(data, errorback) {
        return request._post('/shop/inventory.ErpInventoryRecordOut/list', data, errorback);
    },

    /*出库记录撤销*/
    erpInventoryRecordOutCancel(data, errorback) {
        return request._post('/shop/inventory.ErpInventoryRecordOut/cancel', data, errorback);
    },
    /*出库记录删除*/
    erpInventoryRecordOutDelete(data, errorback) {
        return request._post('/shop/inventory.ErpInventoryRecordOut/delete', data, errorback);
    },


    /*添加报损*/
    erpDamagedProductRecordAdd(data, errorback) {
        return request._post('/shop/inventory.ErpDamagedProductRecord/add', data, errorback);
    },
    
    /*编辑报损*/
    erpDamagedProductRecordUpdate(data, errorback) {
        return request._post('/shop/inventory.ErpDamagedProductRecord/update', data, errorback);
    },

    /*报损审核*/
    erpDamagedProductRecordReview(data, errorback) {
        return request._post('/shop/inventory.ErpDamagedProductRecord/review', data, errorback);
    },

    /*报损删除*/
    erpDamagedProductRecordDelete(data, errorback) {
        return request._post('/shop/inventory.ErpDamagedProductRecord/delete', data, errorback);
    },

    /*获取报损*/
    erpDamagedProductRecordList(data, errorback) {
        return request._post('/shop/inventory.ErpDamagedProductRecord/list', data, errorback);
    },

    /*获取月度报表*/
    getMonthlyStatistics(data, errorback) {
        return request._post('/shop/inventory.ErpInventory/monthlyStatistics', data, errorback);
    },

}

export default InventoryApi;
