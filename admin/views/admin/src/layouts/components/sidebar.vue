<template>
  <div class="overflow-hidden absolute top-16 left-0 bottom-0 bg-white flex">
    <div class="border-r border-[#e4e7ed] w-[200px] h-full flex-shrink-0">
      <div class="p-5 flex flex-col items-center gap-2 overflow-hidden">
        <el-avatar :size="64" :src="settingConfig?.brand_logo" style="--el-avatar-bg-color: #f0f2f5; --el-avatar-text-color: #c0c4cc">
          <ti-icon name="images" size="20" />
        </el-avatar>
        <div class="text-base font-medium truncate text-center">
          <ti-auto-tips :content="settingConfig?.brand_name" placement="bottom" />
        </div>
      </div>
      <ul v-if="!roleLoading" class="py-3">
        <li
          v-for="item in menuList"
          :key="item.path"
          v-permission="item.permission"
          class="cursor-pointer flex items-center gap-2 p-4 text-sm hover:bg-[#fffaeb] relative after"
          :class="{ 'menu-active': isMenuActive(item) }"
          @click="handleToPath(item)"
        >
          <ti-icon :name="item.icon" size="18" />
          <span>{{ item.title }}</span>
        </li>
      </ul>
      <div v-else class="p-3"><el-skeleton :rows="5" animated /></div>
    </div>
    <!-- 二级导航 -->
    <ul v-if="menuChildList && menuChildList.length > 0" class="w-[140px] h-full flex-shrink-0 py-3">
      <template v-if="!roleLoading">
        <li
          v-for="item in menuChildList"
          :key="item.path"
          v-permission="item.permission"
          @click="router.push(item.path)"
          class="cursor-pointer py-3 px-4 hover:bg-[#fffaeb]"
          :class="{ 'text-primary': item.path == route.path }"
        >
          {{ item.title }}
        </li>
      </template>
      <div v-else class="p-3"><el-skeleton :rows="5" animated /></div>
    </ul>
  </div>
</template>

<script setup lang="ts">
  import { watch, computed } from 'vue';
  import { useRouter, useRoute } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import { useUserInfoStore } from '@/stores/userInfo';
  import { $t } from '@/i18n';

  const userInfoStore = useUserInfoStore();
  const { settingConfig, roleLoading } = storeToRefs(userInfoStore);
  const emits = defineEmits<{
    (e: 'change', value?: any[]): void;
  }>();
  const router = useRouter();
  const route = useRoute();
  const menuList = [
    // {
    //   icon: 'dashboard',
    //   title: $t('仪表盘'),
    //   path: '/home',
    // },
    {
      icon: 'set-up',
      title: $t('商家管理'),
      path: '/merchant',
      permission: ['admin_shop_index'],
    },
    {
      icon: 'dashboard',
      title: $t('商家授权管理'),
      path: '/authorization',
      permission: ['admin_erpnext_index'],
    },
    {
      icon: 'store',
      title: $t('用户管理'),
      path: '/user',
      permission: ['admin_admin.user_index', 'admin_admin.role_index', 'admin_admin.loginlog_index', 'admin_admin.optlog_index'],
      children: [
        {
          icon: '',
          title: $t('管理员'),
          path: '/user/admin',
          permission: ['admin_admin.user_index'],
        },
        {
          icon: '',
          title: $t('角色权限'),
          path: '/user/role',
          permission: ['admin_admin.role_index'],
        },
        {
          icon: '',
          title: $t('登录日志'),
          path: '/user/login-log',
          permission: ['admin_admin.loginlog_index'],
        },
        {
          icon: '',
          title: $t('操作日志'),
          path: '/user/operation-log',
          permission: ['admin_admin.optlog_index'],
        },
        {
          icon: '',
          title: $t('人员管理'),
          path: '/user/staff',
          permission: ['admin_admin.staff_index'],
        },
      ],
    },

    {
      icon: 'dashboard',
      title: $t('客户端管理'),
      path: '/client',
      permission: ['admin_client.client_index'],
      children: [
        {
          icon: '',
          title: $t('收银端'),
          path: '/client/cashier',
          permission: ['admin_client.client_index'],
        },
        {
          icon: '',
          title: $t('平板端'),
          path: '/client/slab',
          permission: ['admin_client.client_index'],
        },
        {
          icon: '',
          title: $t('厨显端'),
          path: '/client/kitchen',
          permission: ['admin_client.client_index'],
        },
        {
          icon: '',
          title: $t('商家后台端'),
          path: '/client/manage',
          permission: ['admin_client.client_index'],
        },
        {
          icon: '',
          title: $t('点餐助手'),
          path: '/client/assistant',
          permission: ['admin_client.client_index'],
        },
        {
          icon: '',
          title: $t('自助点餐机'),
          path: '/client/kiosk',
          permission: ['admin_client.client_index'],
        },
      ],
    },
    {
      icon: 'takeaway-fill',
      title: $t('外送管理'),
      path: '/takeout',
      permission: ['admin_client.client_index'],
      children: [
        {
          icon: '',
          title: $t('外送设置'),
          path: '/takeout/setting',
          permission: ['admin_client.client_index'],
        },
        {
          icon: '',
          title: $t('外送商家'),
          path: '/takeout/shop',
          permission: ['admin_client.client_index'],
        },
        {
          icon: '',
          title: $t('外送台单'),
          path: '/takeout/order',
          permission: ['admin_client.client_index'],
        },
      ],
    },
    {
      icon: 'log',
      title: $t('日志管理'),
      path: '/log',
      permission: ['admin_takeout_logs'],
    },
    {
      icon: 'settings',
      title: $t('系统设置'),
      path: '/settings',
      permission: ['admin_setting.service_index'],
    },
  ];
  // 判断一级菜单是否激活
  const isMenuActive = (item: any) => {
    const currentPath = route.path || '';
    const menuPath = item.path || '';

    // 如果没有子菜单，精确匹配
    if (!item.children || item.children.length === 0) {
      return currentPath === menuPath;
    }

    // 如果有子菜单，检查当前路由是否精确匹配一级菜单路径，或者匹配任何子菜单路径
    return currentPath === menuPath || item.children.some((child: any) => currentPath === child.path);
  };

  const menuChildList = computed(() => {
    const activeMenu = menuList.find((item) => isMenuActive(item));
    return activeMenu?.children || [];
  });

  const handleToPath = (item: any) => {
    if (!item.children || item.children.length == 0) {
      router.push(item.path);
    } else if (item.children && item.children[0]) {
      (item.children || []).some((childItem: any) => {
        if (childItem.path && userInfoStore.hasPermission(childItem?.permission)) {
          router.push(childItem.path);
          return true;
        }
      });
    }
  };

  watch(
    () => menuChildList.value,
    (val: any[]) => {
      emits('change', val);
    },
    {
      immediate: true,
    },
  );
</script>

<style lang="scss" scoped>
  .menu-active {
    color: #ffbe00;
    background-color: #fffaeb;

    &::after {
      content: '';
      display: block;
      position: absolute;
      right: 0;
      top: 0;
      width: 4px;
      height: 100%;
      background-color: #ffbe00;
    }
  }
</style>
