import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import { isLogin } from '@/utils/auth';
import { type SettingServiceType, type RoleListData, type versionType, getSettingService, getUserRoleList } from '@/api/setting';
import { flatMap, compact } from 'lodash-es';
import { replaceLinkIcon } from '@/utils';

export const getDeepIds = (arr: RoleListData[] = []): string[] => {
  return flatMap(arr, (item) => {
    return [
      (item.api_path || '')
        .trim()
        .replace(/^\/|\/$/g, '')
        .replaceAll('/', '_'),
      ...getDeepIds(item.children),
    ];
  });
};

export const useUserInfoStore = defineStore('userInfo', () => {
  const userInfo = ref();
  const settingConfig = ref<SettingServiceType>();
  const versionConfig = ref<versionType>();
  //
  const roleLoading = ref(false);
  const roleList = ref<RoleListData[]>();
  const roleListPath = computed(() => compact(getDeepIds(roleList.value)));
  const getUserInfo = (isLoad: boolean = false) => {
    if (!isLogin()) return;
    if (!isLoad && !!userInfo.value.token) return;
    try {
      // const res = await getInfo();
      // userInfo.value = res.data;
    } catch (error) {
      //
    }
  };

  const getSetting = async (isLoad: boolean = false) => {
    if (!isLogin()) return;
    if (!isLoad && !!settingConfig.value) return;
    return new Promise((resolve, reject) => {
      getSettingService()
        .then((res) => {
          settingConfig.value = res.data?.vars?.values || {};
          versionConfig.value = res.data?.vars?.version || {};
          // 存储
          localStorage.setItem('TITLE_BASE', JSON.stringify(settingConfig.value));
          // 设置、替换favicon图标
          if (settingConfig.value?.browser_logo) {
            replaceLinkIcon(settingConfig.value?.browser_logo);
          }
          //
          resolve(res);
        })
        .catch((err) => {
          reject(err);
        });
    });
  };

  // 获取权限列表
  const getRoleList = async (isLoad: boolean = false) => {
    if (!isLogin()) return;
    if (!isLoad && !!roleList.value) return;
    return new Promise((resolve, reject) => {
      roleLoading.value = true;
      getUserRoleList()
        .then((res) => {
          roleList.value = res.data?.menus;
          resolve(res);
        })
        .catch((err) => {
          reject(err);
        })
        .finally(() => {
          roleLoading.value = false;
        });
    });
  };

  // 判断权限
  const hasPermission = (arr: string[] = []) => {
    if (!arr || arr.length == 0) return false;
    return roleListPath.value?.some((role: string) => arr?.includes(role));
  };

  return {
    userInfo,
    getUserInfo,
    settingConfig,
    versionConfig,
    getSetting,
    roleLoading,
    roleList,
    roleListPath,
    getRoleList,
    hasPermission,
  };
});
