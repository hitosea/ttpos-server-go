import { defineStore } from 'pinia';
import { setStorage, getStorage } from '@/utils/storageData';
import { computed } from 'vue';
import AuthApi from '@/api/auth.js';
import configObj from '@/config';
let { strongToken, renderMenu, menu, currency, supplier, erp } = configObj;
import { handRouterTable, handMenuData } from '@/utils/router';
import { EEUIRELOAD } from '@/utils/platform.js';

export const useUserStore = defineStore('main', {
  state: () => {
    return {
      token: getStorage(strongToken),
      userInfo: getStorage('userInfo'),
      list: {},
      menus: getStorage(menu),
      renderMenus: getStorage(renderMenu),
      currency: getStorage(currency),
      supplier: getStorage(supplier),
      erp: getStorage(erp),
    };
  },
  getters: {
    erp_is_open: (state) => state.erp.is_open,
    erp_site_code: (state) => state.erp.site_code,
  },
  actions: {
    setMenus(data) {
      this.menus = data;
      setStorage(JSON.stringify(data), menu);
    },
    setRenderMenus(data) {
      this.renderMenus = data;
      setStorage(JSON.stringify(data), renderMenu);
    },

    computedSupplier() {
      return {
        supplier: computed(() => {
          return getStorage(supplier);
        }),
      };
    },

    computedMenus() {
      return {
        menus: computed(() => {
          return getStorage(menu);
        }),
      };
    },

    computedRenderMenus() {
      return {
        renderMenus: computed(() => {
          return getStorage(renderMenu);
        }),
      };
    },

    computedUserInfo() {
      return {
        userInfo: computed(() => {
          return getStorage('userInfo');
        }),
        currency: computed(() => {
          return getStorage(currency);
        }),
        token: computed(() => {
          return getStorage(strongToken);
        }),
      };
    },

    bus_on(name, fn) {
      let self = this;
      self.list[name] = self.list[name] || [];
      self.list[name].push(fn);
    },
    // 发布
    bus_emit(name, data) {
      let self = this;
      if (self.list[name]) {
        self.list[name].forEach((fn) => {
          fn(data);
        });
      }
    },
    // 取消订阅
    bus_off(name) {
      let self = this;
      if (self.list[name]) {
        delete self.list[name];
      }
    },
    /**
     * @description 登录
     * @param {*} token
     */
    async afterLogin(info) {
      this.userInfo = this.userInfo || {};
      const {
        data: { app_id, shop_supplier_id, supplier_name, token, user_name, user_type, version, currency },
      } = info;
      const {
        data: { menus },
      } = await AuthApi.getRoleList({ token });
      let renderMenusList = handMenuData(JSON.parse(JSON.stringify(menus)));
      let menusList = handRouterTable(JSON.parse(JSON.stringify(menus)));
      //
      let appId = app_id;
      renderMenusList.forEach((item) => {
        item.path = appId + item.path;
        item.redirect_name && (item.redirect_name = '/' + appId + item.redirect_name);
        item.children?.forEach((child) => {
          child.path = '/' + appId + child.path;
          child.redirect_name && (child.redirect_name = '/' + appId + child.redirect_name);
          child.children?.forEach((childItem) => {
            childItem.path = '/' + appId + childItem.path;
            childItem.redirect_name && (childItem.redirect_name = '/' + appId + childItem.redirect_name);
          });
        });
      });
      //
      setStorage(JSON.stringify(menusList), menu);
      setStorage(JSON.stringify(renderMenusList), renderMenu);
      this.userInfo.shop_supplier_id = shop_supplier_id;
      this.userInfo.userName = user_name;
      this.userInfo.version = version;
      this.userInfo.AppID = app_id;
      this.userInfo.supplier_name = supplier_name;
      this.userInfo.user_type = user_type;
      this.token = token;
      this.currency = currency;
      this.renderMenus = renderMenusList;
      if (this.menus && menusList && this.menus.length != menusList.length) {
        this.menus = menusList;
        EEUIRELOAD();
      } else {
        this.menus = menusList;
      }
      setStorage(JSON.stringify(currency), 'currency');
      setStorage(JSON.stringify(token), strongToken);
      setStorage(JSON.stringify(this.userInfo), 'userInfo');
    },

    async changeUserInfo(info) {
      this.userInfo = this.userInfo || {};
      const {
        data: { shop_name, logoUrl, is_open_tax },
      } = info;
      this.userInfo.logoUrl = logoUrl;
      this.userInfo.shopName = shop_name;
      this.userInfo.isOpenTax = is_open_tax;
      setStorage(JSON.stringify(this.userInfo), 'userInfo');
      // 确保异步完成，等待微任务队列执行
      await Promise.resolve();
    },

    /**
     * @description 退出登录
     * @param {*} token
     */
    afterLogout() {
      sessionStorage.clear();
      this.token = null;
      this.menus = null;
      this.userInfo = null;
      setStorage(null, 'userInfo');
    },
    setCurrency(data) {
      this.currency = data;
      setStorage(JSON.stringify(data), 'currency');
    },

    setBaseSale(data) {
      this.baseSale = data;
      setStorage(JSON.stringify(data), 'baseSale');
    },
    /**
     * @description 更改商城设置名称及其logo
     * @param {*} token
     */
    changStore(data) {
      this.userInfo.shopName = data.name;
      this.userInfo.logoUrl = data.logoUrl;
      setStorage(JSON.stringify(this.userInfo), 'userInfo');
    },
  },
});
export default useUserStore;
