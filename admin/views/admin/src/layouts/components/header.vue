<template>
  <div class="h-16 bg-[#100a05] text-white flex items-center justify-between px-4">
    <div class="flex items-center gap-1 w-[180px]">
      <el-image :src="settingConfig?.brand_logo_long" style="width: 146px; height: 40px" class="flex items-center justify-center">
        <template #error>
          <div class="flex gap-1">
            <img class="block" src="@/assets/img/logo.svg" alt="" />
            <img class="block" src="@/assets/img/logo-title.svg" alt="" />
          </div>
        </template>
      </el-image>
    </div>
    <div class="flex items-center gap-6">
      <!-- 语言 -->
      <div class="ti-head-language">
        <ti-language :showText="false" :iconSize="18" />
      </div>
      <!-- 用户信息 -->
      <el-dropdown trigger="hover" @command="handleCommand">
        <div class="cursor-pointer flex items-center justify-center gap-2 text-white">
          <!-- <el-avatar :size="24" :src="userInfo?.avatar">
            <ti-icon name="avatar" />
          </el-avatar> -->
          <span>{{ userInfo?.user_name || '' }}</span>
          <el-icon><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="edit-password">
              <el-icon><Edit /></el-icon>
              <span>{{ $t('修改密码') }}</span>
            </el-dropdown-item>
            <el-dropdown-item command="logout">
              <el-icon><SwitchButton /></el-icon>
              <span>{{ $t('退出登录') }}</span>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
  <!-- 修改密码 -->
  <update-password v-model:show="passwordShow" />
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import { useUserInfoStore } from '@/stores/userInfo';
  import { ElMessageBox } from 'element-plus';
  import { clearAuthInfo, getUserInfo } from '@/utils/auth';
  import { message } from '@/utils/feedback';
  import { logout } from '@/api/login';
  import { $t } from '@/i18n';
  //
  import UpdatePassword from './update-password.vue';

  const { settingConfig } = storeToRefs(useUserInfoStore());
  const route = useRoute();
  const router = useRouter();
  const userInfo = ref(getUserInfo());
  const passwordShow = ref(false);

  const handleCommand = (command: string) => {
    switch (command) {
      case 'edit-password':
        passwordShow.value = true;
        break;
      case 'logout':
        handleLogout();
        break;
      default:
        break;
    }
  };

  const handleLogout = () => {
    ElMessageBox.confirm($t('此操作将退出登录, 是否继续?'), $t('提示'), {
      confirmButtonText: $t('确定'),
      cancelButtonText: $t('取消'),
      type: 'warning',
      beforeClose: async (action: string, instance: any, done: () => void) => {
        if (action === 'confirm') {
          try {
            instance.confirmButtonLoading = true;
            //
            const res = await logout();
            message.success(res.msg);
            clearAuthInfo();
            router.replace(`/login?redirect=${route.fullPath}`);
            //
            done();
          } catch (error) {
            //
          } finally {
            instance.confirmButtonLoading = false;
          }
        } else {
          done();
        }
      },
    })
      .then(() => {})
      .catch(() => {});
  };
</script>

<style lang="scss" scoped>
  .ti-head-language {
    display: flex;
    align-items: center;
    :deep(.el-tooltip__trigger) {
      color: white !important;
    }
  }
</style>
