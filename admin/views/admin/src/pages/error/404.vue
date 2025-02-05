<template>
  <error code="404" :title="$t('哎呀，出错了！您访问的页面不存在...')">
    <template #content>
      <div class="size-40 m-auto">
        <img class="block w-full" src="@/assets/img/403.png" alt="" />
      </div>
    </template>
  </error>
</template>

<script lang="ts" setup>
  import { $t } from '@/i18n';
  import Error from './components/error.vue';
</script>

<style lang="scss" scoped></style>
