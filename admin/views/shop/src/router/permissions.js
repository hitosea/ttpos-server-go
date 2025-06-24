/**
 * @description 路由守卫，目前两种模式：all模式与intelligence模式
 */
import { computed } from 'vue';
import { useUserStore } from '@/store';
import { languageStore } from '@/store/model/language';
import { useLockscreenStore } from '@/store/model/lockscreen';
import dealWithRoute from './dealWithRoute.js';
import tigerImg from '@/assets/TTPOS.ico';

export async function setupPermissions(router) {
  const { bus_emit } = useUserStore();
  const useLockscreen = useLockscreenStore();
  const cloudBasic = languageStore().getCloudBasic().cloudBasic;
  // const useLockscreen = useLockscreenStore();
  // const isLock = computed(() => useLockscreen.isLock);
  let load = 0;
  router.beforeEach(async (to, from, next) => {
    bus_emit('MenuName', to.meta && to.meta.title);
    const { token, menus, computedRenderMenus, computedSupplier, afterLogout } = useUserStore();
    const renderMenus = computedRenderMenus().renderMenus;
    const renderMenusArr = [];
    (renderMenus.value || []).map((item) => {
      renderMenusArr.push(item.path);
      if (item.children) {
        (item.children || []).map((items) => {
          renderMenusArr.push(items.path);
        });
      }
    });
    const supplier = computedSupplier().supplier;
    const app_id = supplier.value?.app_id;
    const expireShow = computed(() => useLockscreen.expire);
    const whiteList = ['/login'];
    //设置logo
    let iconLink = document.createElement('link');
    iconLink.rel = 'icon'; // 设置关系为 icon
    if (cloudBasic.value.base) {
      iconLink.href = cloudBasic.value.base?.browser_logo || '';
      document.title = cloudBasic.value.base?.browser_title + ' - ' + window.$t(to.meta.title || '首页') || '';
    } else {
      iconLink.href = tigerImg;
      document.title = 'TTPOS';
    }
    // 删除当前页面中已有的 logo，如果有的话
    let oldIcon = document.querySelector("link[rel='icon']");
    if (oldIcon) {
      document.head.removeChild(oldIcon);
    }
    // 将新的 logo 添加到 head 标签中
    document.head.appendChild(iconLink);

    //这是独立页面 start
    renderMenusArr.push(`/${app_id}/product/store/product/add`);
    renderMenusArr.push(`/${app_id}/product/store/product/edit`);
    renderMenusArr.push(`/${app_id}/product/store/product/batch`);
    renderMenusArr.push(`/${app_id}/product/store/product/importProduct`);
    renderMenusArr.push(`/${app_id}/supplier/table/table/importQrcode`);
    renderMenusArr.push(`/${app_id}/auth/role/add`);
    renderMenusArr.push(`/${app_id}/auth/role/edit`);
    renderMenusArr.push(`/${app_id}/store/order/detail`);
    renderMenusArr.push(`/${app_id}/store/history_order/detail`);
    renderMenusArr.push(`/${app_id}/store/recharge/detail`);
    renderMenusArr.push(`/${app_id}/marketing/activity/add`);
    renderMenusArr.push(`/${app_id}/marketing/activity/edit`);

    //这是独立页面 end
    if (!token) {
      if (whiteList.includes(to.path)) {
        next();
        return;
      }
      next('/login');
    } else {
      if (app_id && !to.path.includes(app_id) && to.path != '/login') {
        if (app_id) {
          next(`/${app_id}${to.path}`);
        } else {
          afterLogout();
          next(`/login`);
        }
        return;
      }

      // 过期的时候
      if (!app_id && expireShow) {
        next();
        return;
      }

      //正常找不到路由跳转
      if (app_id && renderMenusArr.indexOf(to.path) === -1 && to.matched.length == 0) {
        next({ path: `/${app_id}/home` });
        return;
      }

      //等于home但是ID不对
      const pattern = /\/([^/]+)/g;
      const matches = to.path.match(pattern);
      if (app_id && to.path.includes('/home') && matches.length > 2) {
        next({ path: `/${app_id}/home` });
        return;
      }

      if (to.path == '/login') {
        next({
          path: '/home',
        });
        return;
      }

      if (
        menus.map((h) => h.path).indexOf(to.path.replace(`/${app_id}/`, '/:catchAll(.*)/')) == -1 &&
        menus.map((h) => h.path).indexOf(to.path.replace(`/${app_id}/`, '/')) == -1
      ) {
        next({
          path: renderMenus.value[0]?.redirect_name,
        });
        return;
      }

      // if(isLock.value && to.path != '/lockscreen'){
      //     next({
      // 		path: '/lockscreen'
      // 	});
      // 	return;
      // }
      // if(!isLock.value && to.path == '/lockscreen'){
      //     next({
      // 		path: '/home'
      // 	});
      // 	return;
      // }

      if (menus && load == 0) {
        load++;
        dealWithRoute(menus);
        next({
          ...to,
          replace: true,
        });
        return;
      }
      next();
    }
  });
}
