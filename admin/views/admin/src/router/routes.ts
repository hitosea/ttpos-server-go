import type { RouteRecordRaw } from 'vue-router';
import { $t } from '@/i18n';

import Layout from '@/layouts/default.vue';
import PageLayout from '@/layouts/page-layout.vue';

//
export const LAYOUT = () => Promise.resolve(Layout);
export const PAGE_LAYOUT = () => Promise.resolve(PageLayout);

// 不需要登录就会注册的路由
export const constantRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    component: () => import('@/pages/login/index.vue'),
    meta: {
      title: $t('登录'),
    },
  },
  {
    path: '/403',
    component: () => import('@/pages/error/403.vue'),
    meta: {
      title: 403,
    },
  },
  {
    path: '/500',
    component: () => import('@/pages/error/500.vue'),
    meta: {
      title: 500,
    },
  },
  {
    path: '/:pathMatch(.*)*',
    component: () => import('@/pages/error/404.vue'),
    meta: {
      title: 404,
    },
  },
];

// 需要登录权限才可以访问的路由
export const asyncRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: LAYOUT,
    redirect: '/merchant',
    // meta: {
    //   title: $t('首页'),
    // },
    children: [
      {
        path: '/home',
        component: () => import('@/pages/home/index.vue'),
        meta: {
          title: $t('仪表盘'),
        },
      },
      {
        path: '/merchant',
        component: () => import('@/pages/merchant/index.vue'),
        meta: {
          title: $t('商家管理'),
        },
      },
      {
        path: '/authorization',
        component: () => import('@/pages/authorization/index.vue'),
        meta: {
          title: $t('商家授权管理'),
        },
      },
      {
        path: '/user',
        redirect: { path: '/user/admin' }, // 重定向
        meta: {
          title: $t('用戶管理'),
        },
        children: [
          {
            path: 'admin',
            component: () => import('@/pages/user/admin.vue'),
            meta: {
              title: $t('管理员'),
            },
          },
          {
            path: 'role',
            component: () => import('@/pages/user/role.vue'),
            meta: {
              title: $t('角色权限'),
            },
          },
          {
            path: 'login-log',
            component: () => import('@/pages/user/login-log.vue'),
            meta: {
              title: $t('登录日志'),
            },
          },
          {
            path: 'operation-log',
            component: () => import('@/pages/user/operation-log.vue'),
            meta: {
              title: $t('操作日志'),
            },
          },
          {
            path: 'staff',
            component: () => import('@/pages/user/staff.vue'),
            meta: {
              title: $t('人员管理'),
            },
          },
        ],
      },
      {
        path: '/client',
        redirect: { path: '/client/cashier' }, // 重定向
        meta: {
          title: $t('客户端管理'),
        },
        children: [
          {
            path: 'cashier',
            component: () => import('@/pages/client/cashier.vue'),
            meta: {
              title: $t('收银端'),
            },
          },
          {
            path: 'slab',
            component: () => import('@/pages/client/slab.vue'),
            meta: {
              title: $t('平板端'),
            },
          },
          {
            path: 'kitchen',
            component: () => import('@/pages/client/kitchen.vue'),
            meta: {
              title: $t('厨显端'),
            },
          },
          {
            path: 'manage',
            component: () => import('@/pages/client/manage.vue'),
            meta: {
              title: $t('商家后台端'),
            },
          },
          {
            path: 'assistant',
            component: () => import('@/pages/client/assistant.vue'),
            meta: {
              title: $t('点餐助手'),
            },
          },
          {
            path: 'kiosk',
            component: () => import('@/pages/client/kiosk.vue'),
            meta: {
              title: $t('自助点餐机'),
            },
          },
        ],
      },
      {
        path: '/takeout',
        redirect: { path: '/takeout/setting' }, // 重定向
        meta: {
          title: $t('外送管理'),
        },
        children: [
          {
            path: 'setting',
            component: () => import('@/pages/takeout/setting.vue'),
            meta: {
              title: $t('外送设置'),
            },
          },
          {
            path: 'shop',
            component: () => import('@/pages/takeout/shop.vue'),
            meta: {
              title: $t('外送商家'),
            },
          },
          {
            path: 'order',
            component: () => import('@/pages/takeout/order.vue'),
            meta: {
              title: $t('外送台单'),
            },
          },
        ],
      },
      {
        path: '/log',
        component: () => import('@/pages/log/grab.vue'),
        meta: {
          title: $t('导入日志'),
        },
      },
      {
        path: '/settings',
        component: () => import('@/pages/settings/index.vue'),
        meta: {
          title: $t('系统设置'),
        },
      },
    ],
  },
];

//
export default [...constantRoutes, ...asyncRoutes];
