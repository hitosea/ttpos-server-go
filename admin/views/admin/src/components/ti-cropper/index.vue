<template>
  <el-dialog :width="props.dialogWidth" :modelValue="props.show" align-center :title="$t('图片裁剪')" @close="emits('update:show', false)">
    <div class="flex items-center justify-center">
      <img class="block w-full" ref="imageRef" :src="imageSrc" alt="" />
    </div>
    <template #footer>
      <el-button @click="emits('update:show', false)"> {{ $t('取消') }}</el-button>
      <el-button type="primary" @click="handelCropper"> {{ $t('确定') }} </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
  import { ref, watch, nextTick } from 'vue';
  import Cropper from 'cropperjs';
  import 'cropperjs/dist/cropper.css';
  import { $t } from '@/i18n';

  const emits = defineEmits<{
    (e: 'update:show', value?: boolean): void;
    (e: 'change', value?: File): void;
  }>();
  const props = withDefaults(
    defineProps<{
      show: boolean;
      imageSrc?: string;
      aspectRatio?: number; // 置裁剪框为固定的宽高比
      viewMode?: Cropper.ViewMode; // 视图控制
      autoCropArea?: number; // 设置裁剪区域占图片的大小 值为 0-1 默认 0.8 表示 80%的区域
      imageName?: string;
      dialogWidth?: number;
    }>(),
    {
      show: false,
      imageSrc: undefined,
      aspectRatio: 1,
      viewMode: 1,
      autoCropArea: 1,
      imageName: Date.now().toString(),
      dialogWidth: 640,
    },
  );
  const imageRef = ref();
  const cropperEl = ref<Cropper>();

  const base64ToFile = (data?: string) => {
    if (!data) return;
    const arr = data.split(',');
    if (!arr || arr.length == 0 || !arr[0]) return;
    const type = arr[0].match(/:(.*?);/);
    if (!type) return;
    const mime = type[1];
    if (!mime) return;
    const suffix = mime.split('/')[1];
    //
    const bstr = atob(arr[1]);
    let n = bstr.length;
    let u8arr = new Uint8Array(n);
    while (n--) {
      u8arr[n] = bstr.charCodeAt(n);
    }
    return new File([u8arr], `${props.imageName.slice(0, props.imageName.lastIndexOf('.'))}.${suffix}`, {
      type: mime,
    });
  };

  const handelCropper = () => {
    const canvas = cropperEl.value?.getCroppedCanvas();
    const croppedImage = canvas?.toDataURL();
    //

    emits('change', base64ToFile(croppedImage));
    emits('update:show', false);
  };

  watch(
    () => props.show,
    () => {
      // 使用Cropper构造函数创建裁剪器实例，并将图片元素和一些裁剪选项传入
      if (props.show) {
        nextTick(() => {
          cropperEl.value = new Cropper(imageRef.value, {
            aspectRatio: props.aspectRatio,
            viewMode: props.viewMode,
            autoCropArea: props.autoCropArea,
          });
        });
      } else {
        cropperEl.value?.destroy();
      }
    },
  );
</script>

<style lang="scss" scoped></style>
