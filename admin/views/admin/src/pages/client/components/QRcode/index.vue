<template>
  <el-dialog width="520" :title="$t('下载二维码')" :modelValue="props.show" :close-on-click-modal="false" :close-on-press-escape="false" align-center @close="handleClose()">
    <div class="max-h-[75vh] overflow-auto pr-4 flex items-center">
      <img :src="url" class="w-[200px] m-auto my-[20px]" />
    </div>
    <template #footer>
      <el-button @click="handleClose()">{{ $t('取消') }}</el-button>
      <el-button type="primary" @click="handleDownload()">{{ $t('下载') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
  import { watch, ref } from 'vue';
  import { qrcodeClient } from '@/api/client';
  const props = withDefaults(
    defineProps<{
      show?: boolean;
      uuid?: string;
    }>(),
    {
      show: false,
      uuid: '',
    },
  );
  const url = ref('');

  const emits = defineEmits<{
    (e: 'update:show', value: boolean): void;
  }>();

  const handleDownload = () => {
    const base64Data = url.value;
    const byteCharacters = atob(base64Data.split(',')[1]);
    const byteNumbers = new Array(byteCharacters.length);

    for (let i = 0; i < byteCharacters.length; i++) {
      byteNumbers[i] = byteCharacters.charCodeAt(i);
    }

    const byteArray = new Uint8Array(byteNumbers);
    const blob = new Blob([byteArray], { type: 'image/png' });

    // 创建一个下载链接并执行下载操作
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = 'image.png';
    a.click();
  };

  const handleClose = () => {
    emits('update:show', false);
  };

  const getCode = async () => {
    try {
      const data = { uuid: '' };
      data.uuid = props.uuid || '';
      const res = await qrcodeClient(data);
      url.value = res.data;
    } catch (error) {
      //
    } finally {
    }
  };

  watch(
    () => props.uuid,
    (val) => {
      if (!val) return;
      getCode();
    },
  );
</script>
<style lang="scss" scoped></style>
