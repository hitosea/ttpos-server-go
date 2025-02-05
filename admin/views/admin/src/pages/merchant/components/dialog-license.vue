<template>
  <el-dialog width="640" :title="$t('授权码')" :modelValue="props.show" :close-on-click-modal="false" :close-on-press-escape="false" @close="emits('update:show', false)">
    <el-form class="ti-license-form" label-position="top" label-width="auto">
      <el-form-item :label="$t('商家：')">
        <div class="w-full overflow-hidden">{{ detailLicense?.data?.name || detailLicense?.data?.supplier_name }}</div>
      </el-form-item>
      <el-form-item :label="$t('授权码：')">
        <div class="w-full overflow-hidden rounded py-2 px-3 bg-[#f6f8fb]">
          <span>{{ detailLicense?.key }}</span>
          <el-button icon="CopyDocument" link style="margin-left: 6px; vertical-align: -2px" @click="handleCopy"></el-button>
          <el-button type="primary" :loading="loading" style="margin-left: 6px; vertical-align: 0px" @click="handleUpdateLicense">{{ $t('更新授权码') }}</el-button>
        </div>
        <div class="text-[#ccc] w-full">{{ $t('更新后商家需重新授权获取最新配置') }}</div>
      </el-form-item>
      <el-form-item>
        <el-alert :type="detailLicense?.expired_time == 0 ? 'success' : 'warning'" :closable="false">
          <template #title>
            <div class="flex gap-1 flex-col">
              <div v-if="detailLicense?.expired_time">
                <span>{{ $t('授权天数：') }}</span>
                <span>{{ detailLicense?.auth_day }}</span>
              </div>
              <div>
                <span>{{ $t('上一次授权码更新时间：') }}</span>
                <span>{{ detailLicense?.update_time || '-' }}</span>
              </div>
              <div>
                <span>{{ $t('当前工控机MAC地址：') }}</span>
                <span>{{ detailLicense?.data?._mac || '-' }}</span>
              </div>
              <div>
                <span>{{ $t('当前工控机序号：') }}</span>
                <span>{{ detailLicense?.data?._uuid || '-' }}</span>
              </div>
              <div>
                <span>{{ $t('过期时间：') }}</span>
                <span>{{ detailLicense?.expired_time || $t('永不过期') }}</span>
              </div>
            </div>
          </template>
        </el-alert>
        <div> </div>
      </el-form-item>
    </el-form>
  </el-dialog>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue';
  import copy from 'copy-to-clipboard';
  import { ShopListData, fetchUpdateLicense } from '@/api/merchant';
  import { message } from '@/utils/feedback';
  import { $t } from '@/i18n';

  const emits = defineEmits<{
    (e: 'update:show', value: boolean): void;
  }>();
  const props = withDefaults(
    defineProps<{
      show?: boolean;
      detail?: ShopListData;
    }>(),
    {
      show: false,
      detail: () => ({}),
    },
  );
  const loading = ref(false);
  const detailLicense = ref<any>({});

  const handleCopy = () => {
    copy(detailLicense.value?.key);
    message.success($t('复制成功'));
  };

  const handleUpdateLicense = async () => {
    try {
      loading.value = true;
      const res = await fetchUpdateLicense(props.detail?.app_id);
      detailLicense.value = res.data?.license;
      message.success(res.msg);
    } catch (error) {
      //
    } finally {
      loading.value = false;
    }
  };

  watch(
    () => props.show,
    (val) => {
      if (!val) return;
      detailLicense.value = props.detail;
    },
  );
</script>

<style lang="scss" scoped></style>
