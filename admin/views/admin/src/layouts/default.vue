<template>
  <div class="bg-[#f6f8fb] absolute inset-0">
    <layout-header />
    <layout-sidebar @change="handleMenuChildList" />
    <div class="absolute top-16 bottom-0 right-0 py-5 px-4 overflow-auto" :class="hasMenuChild ? 'left-[340px]' : 'left-[200px]'">
      <router-view v-if="!roleLoading" v-slot="{ Component, route }">
        <component :is="Component" :key="route.path" />
      </router-view>
      <div v-else class="p-4 bg-white rounded h-full"><el-skeleton :rows="10" animated /></div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { storeToRefs } from 'pinia';
  import { useUserInfoStore } from '@/stores/userInfo';
  //
  import LayoutHeader from './components/header.vue';
  import LayoutSidebar from './components/sidebar.vue';

  const { roleLoading } = storeToRefs(useUserInfoStore());
  const hasMenuChild = ref(false);

  const handleMenuChildList = (list: any) => {
    if (list && list.length > 0) {
      hasMenuChild.value = true;
    } else {
      hasMenuChild.value = false;
    }
  };
</script>

<style lang="scss" scoped></style>
