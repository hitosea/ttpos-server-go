<template>
  <error code="500" :title="$t('对不起，服务器错误！')">
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
