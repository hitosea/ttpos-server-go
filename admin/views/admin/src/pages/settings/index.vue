<template>
  <div class="p-4 bg-white rounded min-h-full">
    <el-form :model="formData" :rules="formRules" ref="formElement" label-position="top" label-width="124px">
      <el-form-item :label="$t('品牌名称：')" prop="brand_name">
        <el-input v-model="formData.brand_name" class="max-width" :controls="false" :precision="0" :min="0" :placeholder="$t('请输入品牌名称')" clearable></el-input>
        <div class="text-[#ccc] w-full">{{ $t('显示在收银机副屏') }}</div>
      </el-form-item>
      <el-form-item :label="$t('品牌LOGO：') + '\n' + $t('(正方型)')" prop="brand_logo">
        <ti-upload v-model="formData.brand_logo" />
        <div class="text-[#ccc] w-full">{{ $t('支持JPG、JPEG、PNG格式，小於15MB，尺寸：64*64px') }}</div>
      </el-form-item>
      <el-form-item :label="$t('品牌LOGO：') + '\n' + $t('(长方型)')" prop="brand_logo_long">
        <ti-upload v-model="formData.brand_logo_long" :aspectRatio="3.65" />
        <div class="text-[#ccc] w-full">{{ $t('支持JPG、JPEG、PNG格式，小於15MB，尺寸：146*40px') }}</div>
      </el-form-item>
      <el-form-item :label="$t('浏览器title：') + '\n' + '（LOGO）'" prop="browser_logo">
        <ti-upload v-model="formData.browser_logo" />
        <div class="text-[#ccc] w-full">{{ $t('支持JPG、JPEG、PNG格式，小於15MB，尺寸：64*64px') }}</div>
      </el-form-item>
      <el-form-item :label="$t('浏览器title：') + '\n' + $t('(品牌名称)')" prop="browser_title">
        <el-input type="text" v-model="formData.browser_title" maxlength="10" :placeholder="$t('名称')"></el-input>
        <div class="text-[#ccc] w-full">{{ $t('必须为10个字符以内') }}</div>
      </el-form-item>
      <el-form-item :label="$t('准备到期提醒：')" prop="expiration_reminder">
        <el-input-number
          v-model="formData.expiration_reminder"
          class="max-width"
          :controls="false"
          :precision="0"
          :min="0"
          :placeholder="$t('请输入准备到期提醒')"
          clearable
        ></el-input-number>
        <div class="px-2 text-[#495060]">{{ $t('天') }}</div>
        <div class="text-[#ccc] w-full">{{ $t('根据填写的时间提醒商家剩余可用时间') }}</div>
      </el-form-item>

      <el-form-item :label="$t('客户端版本：')">
        <p>{{ versionData.client_version }}</p>
      </el-form-item>
      <el-form-item :label="$t('服务端版本：')">
        <p>{{ versionData.server_version }}</p>
      </el-form-item>
      <!-- <el-form-item :label="$t('会员默认头像：')" prop="member_default_avatar">
        <ti-upload v-model="formData.member_default_avatar" />
        <div class="text-[#ccc] w-full">{{ $t('支持JPG、JPEG、PNG格式，小於15MB，尺寸：48*48px') }}</div>
      </el-form-item> -->
      <div class="border-t border-[#eee] flex items-center justify-center p-4">
        <el-button @click="handleReset">{{ $t('重置') }}</el-button>
        <el-button :loading="formLoading" type="primary" @click="handleSubmit">{{ $t('保存') }}</el-button>
      </div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, watch } from 'vue';
  import { storeToRefs } from 'pinia';
  import { message } from '@/utils/feedback';
  import { useUserInfoStore } from '@/stores/userInfo';
  import type { FormRules } from 'element-plus';
  import { type SettingServiceType, type versionType, fetchSettingService } from '@/api/setting';
  import { $t } from '@/i18n';

  const userInfoStore = useUserInfoStore();
  const { settingConfig, versionConfig } = storeToRefs(userInfoStore);
  const formElement = ref();
  const formLoading = ref(false);
  const formData = ref<SettingServiceType>({
    brand_name: '', // 品牌名称
    brand_logo: '', // 品牌LOGO（正方型）
    brand_logo_long: '', // 品牌LOGO（长方型）
    expiration_reminder: 0, // 到期提醒 0为永久
    member_default_avatar: '', // 会员默认头像
    browser_title: '', // title
    browser_logo: '', //titlelogo
  });
  const versionData = ref<versionType>({
    client_version: '', // 客户端号
    server_version: '', // 服务器号
  });
  const formRules = reactive<FormRules>({
    brand_name: [
      {
        required: true,
        message: $t('请输入品牌名称'),
        trigger: 'change',
      },
    ],
    brand_logo: [
      {
        required: true,
        message: $t('请上传品牌LOGO'),
        trigger: 'change',
      },
    ],
    brand_logo_long: [
      {
        required: true,
        message: $t('请上传品牌LOGO'),
        trigger: 'change',
      },
    ],
    browser_logo: [
      {
        required: true,
        message: $t('请上传浏览器title-LOGO'),
        trigger: 'change',
      },
    ],
    browser_title: [
      {
        required: true,
        message: $t('请上传浏览器title'),
        trigger: 'change',
      },
    ],
    // expiration_reminder: [
    //   {
    //     required: true,
    //     message: $t('请输入准备到期提醒'),
    //     trigger: 'change',
    //   },
    // ],
    // member_default_avatar: [
    //   {
    //     required: true,
    //     message: $t('请上传会员默认头像'),
    //     trigger: 'change',
    //   },
    // ],
  });

  const handleSubmit = async () => {
    formElement.value?.validate(async (valid: boolean) => {
      if (!valid) return;
      try {
        formLoading.value = true;
        const res = await fetchSettingService(formData.value);
        message.success(res.msg);
        //
        userInfoStore.getSetting(true);
      } catch (error) {
        //
      } finally {
        formLoading.value = false;
      }
    });
  };

  const handleReset = () => {
    formData.value = { ...settingConfig.value };
    versionData.value = { ...versionConfig.value };
  };

  watch(
    () => settingConfig.value,
    () => {
      handleReset();
    },
    {
      deep: true,
      immediate: true,
    },
  );
</script>

<style lang="scss" scoped></style>
