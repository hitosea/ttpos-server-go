import request from '@/utils/request'

let Terminal = {
    /*获取收银机设置*/
    getTerminal(data, errorback) {
        return request._get('/shop/setting.Terminal/cashier', data, errorback);
    },
    /*收银机设置密码*/
    editCashierPassword(data, errorback) {
        return request._post('/shop/setting.Terminal/editCashierPassword', data, errorback);
    },
    /*收银机锁屏设置密码*/
    editCashierLockPassword(data, errorback) {
        return request._post('/shop/setting.Terminal/editCashierLockPassword', data, errorback);
    },
    /*收银机高级设置密码*/
    editCashierAdvancedPassword(data, errorback) {
        return request._post('/shop/setting.Terminal/editCashierAdvancedPassword', data, errorback);
    },
    /*设置收银机设置*/
    saveTerminal(data, errorback) {
        return request._post('/shop/setting.Terminal/cashier', data, errorback);
    },

    /*获取平板端设置*/
    getTablet(data, errorback) {
        return request._get('/shop/setting.Terminal/tablet', data, errorback);
    },
    /*平板端设置密码*/
    editAdvancedPassword(data, errorback) {
        return request._post('/shop/setting.Terminal/editAdvancedPassword', data, errorback);
    },
    /*设置平板端设置*/
    saveTablet(data, errorback) {
        return request._post('/shop/setting.Terminal/tablet', data, errorback);
    },

    /*获取厨显设置*/
    getKitchen(data, errorback) {
        return request._get('/shop/setting.Terminal/kitchen', data, errorback);
    },
    /*设置厨显设置*/
    saveKitchen(data, errorback) {
        return request._post('/shop/setting.Terminal/kitchen', data, errorback);
    },
    /*厨显设置密码*/
    editKitchenAdvancedPassword(data, errorback) {
        return request._post('/shop/setting.Terminal/editKitchenAdvancedPassword', data, errorback);
    },

    /*获取点餐助手设置*/
    getAssistant(data, errorback) {
        return request._get('/shop/setting.Terminal/assistant', data, errorback);
    },
    /*设置厨显设置*/
    saveAssistant(data, errorback) {
        return request._post('/shop/setting.Terminal/assistant', data, errorback);
    },
    /*设置点餐助手高级密码*/
    editAssistantAdvancedPassword(data, errorback) {
        return request._post('/shop/setting.Terminal/editAssistantAdvancedPassword', data, errorback);
    },
    /*设置点餐助手锁屏密码*/
    editAssistantLockPassword(data, errorback) {
        return request._post('/shop/setting.Terminal/editAssistantLockPassword', data, errorback);
    },

    /*获取H5设置*/
    getH5(data, errorback) {
        return request._get('/shop/setting.Terminal/h5', data, errorback);
    },
    /*设置H5设置*/
    saveH5(data, errorback) {
        return request._post('/shop/setting.Terminal/h5', data, errorback);
    },
}
export default Terminal;