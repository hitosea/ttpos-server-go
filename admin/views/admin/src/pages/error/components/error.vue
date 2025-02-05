<template>
  <div class="absolute text-center size-full flex items-center justify-center">
    <div>
      <slot name="content">
        <div class="text-[#100a05] text-[150px]">{{ code }}</div>
      </slot>
      <div class="text-lg text-[#100a05] m-7">{{ title }}</div>
      <el-button v-if="props.showBtn" type="primary" @click="router.go(-1)">{{ $t('返回上一页') }}</el-button>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { useRouter } from 'vue-router';
  import { $t } from '@/i18n';

  const props = withDefaults(
    defineProps<{
      code: string;
      title: string;
      showBtn?: boolean;
    }>(),
    {
      showBtn: true,
    },
  );

  const router = useRouter();
</script>

<style lang="scss" scoped></style>
