<template>
  <div class="flex items-center">
    <p class="mr-[12px]" v-if="!loading">{{ fileName }}</p>
    <el-upload class="" :accept="props.accept" :auto-upload="false" :show-file-list="false" :multiple="props.multiple" :on-change="handleUpload">
      <el-button type="primary" :loading="loading">{{ $t('上传') }}</el-button>
    </el-upload>
  </div>
</template>

<script setup lang="ts">
  import { uploadFile } from '@/api/file';
  import { message } from '@/utils/feedback';
  import { $t } from '@/i18n';
  import { ref } from 'vue';

  const fileName = ref('');
  const loading = ref(false);

  const emits = defineEmits<{
    (e: 'update:modelValue', value?: string): void;
    (e: 'change', value?: string): void;
  }>();
  const props = withDefaults(
    defineProps<{
      modelValue?: string;
      size?: number;
      accept?: string;
      multiple?: boolean;
    }>(),
    {
      modelValue: undefined,
      size: 100,
      accept: 'application/vnd.android.package-archive', // image/jpeg,image/png,image/jpg,video/*
      multiple: false,
    },
  );

  const handleUpload = async (file: any) => {
    if (file.size > props.size * 1024 * 1024) {
      message.warning($t('安装包大小不能超过100MB！'));
      return false;
    }

    const reader = new FileReader();
    reader.onload = async (event) => {
      const binaryData = event.target?.result || file; // 获取二进制数据
      // 创建一个新 Blob 对象
      const binaryBlob = new Blob([binaryData], { type: file.raw.type });
      const date = Number(new Date());
      const files = new File([binaryBlob], file.name, { lastModified: date, type: binaryBlob.type });
      loading.value = true;
      try {
        const data = new FormData();
        data.append('file', files);
        const res = await uploadFile(data);
        fileName.value = file.name;
        message.success($t('上传成功'));
        loading.value = false;
        emits('change', res.data);
        emits('update:modelValue', res.data?.file_path);
      } catch (error) {
        //
        loading.value = false;
      }
    };
    // 读取 File 对象内容并转换为二进制数据
    reader.readAsArrayBuffer(file.raw);
  };

  const clearName = () => {
    fileName.value = '';
  };

  defineExpose({
    clearName,
  });
</script>

<style lang="scss" scoped></style>
