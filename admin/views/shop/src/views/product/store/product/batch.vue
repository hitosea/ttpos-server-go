<template>
  <batchCompoent v-if="type != '5'"></batchCompoent>
  <importProduct v-if="type == '5'"></importProduct>
  <batchDiscount v-if="type == '6'"></batchDiscount>
</template>

<script setup>
  import { ref, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import batchCompoent from './batchCompoent.vue';
  import importProduct from './importProduct.vue';
  import batchDiscount from './batchDiscount.vue';

  // 获取路由实例
  const route = useRoute();

  // 响应式数据
  const type = ref('');
  const title = ref('');

  // 组件挂载时初始化
  onMounted(() => {
    type.value = route.query.type;
    title.value = route.query.title;
  });
</script>

<style lang=""></style>
