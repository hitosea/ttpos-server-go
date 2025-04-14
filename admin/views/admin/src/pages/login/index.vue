<template>
  <div class="ti-login bg-[#fff6de] absolute inset-0 flex items-center justify-center overflow-hidden">
    <!-- 表格 -->
    <div class="rounded-md bg-white py-14 px-10 max-w-[480px]">
      <div class="flex items-center justify-center gap-3 mb-10">
        <span class="text-2xl font-extrabold truncate text-primary">
          <template v-if="brand.value == 'JBC'">JBCレジ</template>
          <template v-if="brand.value == 'TTPOS'">TTPOS</template>
          ·
          {{ $t('平台管理') }}
        </span>
      </div>
      <el-form ref="formElement" :model="formData" :rules="formRules" label-position="left" size="large" style="--el-component-size-large: 48px">
        <el-form-item prop="username">
          <el-input type="text" v-model="formData.username" maxlength="50" :placeholder="$t('请输入用户名')" @keyup.enter="handleSubmit">
            <template #prefix>
              <ti-icon size="20" name="user" color="#100A05" />
            </template>
          </el-input>
        </el-form-item>
        <el-form-item prop="password">
          <el-input type="password" v-model="formData.password" maxlength="50" :placeholder="$t('请输入登录密码')" @input="handleInput" show-password @keyup.enter="handleSubmit">
            <template #prefix>
              <ti-icon size="20" name="lock" color="#100A05" />
            </template>
          </el-input>
        </el-form-item>
        <el-form-item class="ti-verify-code-box" prop="code">
          <el-input class="" type="text" v-model="formData.code" maxlength="50" :placeholder="$t('请输入验证码')" @keyup.enter="handleSubmit">
            <template #prefix>
              <ti-icon size="20" name="verify-code" color="#100A05" />
            </template>
          </el-input>
          <div class="rounded bg-[#e4e7ed] w-[194px] h-[48px]" @click="handleGetCaptcha">
            <img v-if="captchaImg" class="block h-full w-full" :src="captchaImg" alt="" />
            <ti-loading v-else-if="captchaLoading" />
            <view v-else class="ti-login-code-error"> </view>
          </div>
        </el-form-item>
        <div class="mt-2">
          <el-button
            class="w-full"
            type="primary"
            :loading="formLoading"
            :disabled="formDisabled"
            style="--el-button-size: 48px; --el-font-size-base: 16px; font-weight: 500"
            @click="handleSubmit"
          >
            {{ $t('登录') }}
          </el-button>
        </div>
      </el-form>
    </div>
    <!-- 切换语言 -->
    <div class="absolute bottom-8 left-1/2 -translate-x-1/2">
      <ti-language />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick } from 'vue';
  import { useRouter, useRoute } from 'vue-router';
  import { message } from '@/utils/feedback';
  import type { FormRules } from 'element-plus';
  import { login, getCaptcha, getService } from '@/api/login';
  import { setToken, setUserInfo } from '@/utils/auth';
  import { useUserInfoStore } from '@/stores/userInfo';
  import { $t } from '@/i18n';
  import { replaceLinkIcon } from '@/utils';

  const { getSetting, getRoleList } = useUserInfoStore();
  const route = useRoute();
  const router = useRouter();
  const formLoading = ref(false);
  const formElement = ref();
  const formData = ref({
    username: '',
    password: '',
    code: '',
    sign: '',
  });
  const brand = ref<any>('JBC');
  if ((window as any).config instanceof Function) {
    brand.value = ref((window as any).config().brand.toUpperCase());
  }
  const formRules = reactive<FormRules>({
    username: [
      {
        required: true,
        message: $t('请输入用户名'),
        trigger: ['change'],
      },
    ],
    password: [
      {
        required: true,
        message: $t('请输入登录密码'),
        trigger: ['change'],
      },
    ],
    code: [
      {
        required: true,
        message: $t('请输入验证码'),
        trigger: ['change'],
      },
    ],
  });
  const formDisabled = computed(() => !formData.value.username || !formData.value.password || !formData.value.code);
  const captchaImg = ref();
  const captchaLoading = ref(false);

  const handleSubmit = () => {
    formElement.value?.validate(async (valid: boolean) => {
      if (!valid) return;
      try {
        formLoading.value = true;
        const res = await login(formData.value);
        message.success(res.msg);
        if (res.data?.token) {
          // 设置token
          setToken(res.data.token);
          setUserInfo(res.data);
          // 权限信息
          await getRoleList(true);
          // 基础信息
          await getSetting(true);
          // 跳转到首页
          router.replace((route.query?.redirect as string) || '/');
        }
      } catch (error) {
        handleGetCaptcha();
      } finally {
        formLoading.value = false;
      }
    });
  };

  const handleGetCaptcha = async () => {
    if (captchaLoading.value) return;
    try {
      captchaLoading.value = true;
      formLoading.value = true;
      const form: any = { v: 1 };
      const res = await getCaptcha(form);
      captchaImg.value = res.data.base64;
      formData.value.sign = res.data.sign;
    } catch (error) {
      console.log(error);
    } finally {
      captchaLoading.value = false;
      formLoading.value = false;
    }
  };

  const handleVisibilityChange = async () => {
    handleGetCaptcha();
  };

  const handleInput = () => {
    //过滤密码中的空格符号
    nextTick(() => {
      formData.value.password = formData.value.password.replace(/\s/g, '');
    });
  };

  const handleGetService = async () => {
    const res = await getService();
    // 设置、替换favicon图标
    if (res.data?.base?.browser_logo) {
      replaceLinkIcon(res.data?.base?.browser_logo);
    }
  };

  onMounted(() => {
    document.addEventListener('visibilitychange', handleVisibilityChange);
  });

  onBeforeUnmount(() => {
    document.addEventListener('visibilitychange', handleVisibilityChange);
  });

  handleGetCaptcha();
  handleGetService();
</script>

<style lang="scss" scoped>
  .ti-verify-code-box {
    :deep(.el-form-item__content) {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
      .el-input {
        flex: 1 1 0%;
      }
    }
  }
  .ti-login {
    background-image: url('@/assets/img/login/bg.png');
    background-position: right bottom;
    background-repeat: no-repeat;
  }
  .ti-login-code-error {
    display: flex;
    align-items: center;
    justify-content: center;
    color: #999;
    height: 100%;
  }
</style>
