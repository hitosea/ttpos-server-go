<template>
  <error code="403" :title="$t('您的账号权限不足，请联系管理员添加权限！')">
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
