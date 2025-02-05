<template>
  <div>
    <el-upload
      class="size-12 block border border-dashed border-[#eee]"
      :accept="props.accept"
      :auto-upload="false"
      :show-file-list="false"
      :multiple="props.multiple"
      :before-upload="handleBeforeUpload"
      :on-change="handleCropper"
    >
      <img v-if="props.modelValue" :src="props.modelValue" class="block size-full" />
      <el-icon v-else class="text-lg size-12 text-center"><Plus /></el-icon>
    </el-upload>
    <!-- 图片裁剪 -->
    <ti-cropper v-model:show="cropperShow" :aspectRatio="props.aspectRatio" :imageName="cropperName" :imageSrc="cropperImg" @change="handleUploadImage" />
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { fetchUploadFile } from '@/api/file';
  import { message } from '@/utils/feedback';
  import { $t } from '@/i18n';

  const emits = defineEmits<{
    (e: 'update:modelValue', value?: string): void;
    (e: 'change', value?: string): void;
  }>();
  const props = withDefaults(
    defineProps<{
      modelValue?: string;
      imgSize?: number;
      videoSize?: number;
      accept?: string;
      multiple?: boolean;
      aspectRatio?: number;
    }>(),
    {
      modelValue: undefined,
      imgSize: 15,
      videoSize: 30,
      accept: 'image/jpeg,image/png,image/jpg', // image/jpeg,image/png,image/jpg,video/*
      multiple: false,
      aspectRatio: 1,
    },
  );
  const cropperName = ref();
  const cropperImg = ref();
  const cropperShow = ref();

  const handleBeforeUpload = (file: File) => {
    if (file.size > props.imgSize * 1024 * 1024 && file.type.includes('image')) {
      message.warning($t('图片大小不能超过{count}MB！', { count: props.imgSize }));
      return false;
    }
    if (file.size > props.videoSize * 1024 * 1024 && file.type.includes('video')) {
      message.warning($t('视频大小不能超过{count}MB！', { count: props.videoSize }));
      return false;
    }
    return true;
  };

  const handleCropper = (e: any) => {
    if (!handleBeforeUpload(e.raw)) return;
    //
    cropperName.value = e.raw.name;
    //
    const reader = new FileReader();
    reader.onloadend = () => {
      cropperImg.value = reader.result;
      cropperShow.value = true;
    };
    reader.readAsDataURL(e.raw);
  };

  const handleUploadImage = async (file: File) => {
    try {
      const data = new FormData();
      data.append('iFile', file);
      const res = await fetchUploadFile(data);
      emits('update:modelValue', res.data?.file_path);
      emits('change', res.data?.file_path);
    } catch (error) {
      //
    }
  };
</script>

<style lang="scss" scoped></style>
