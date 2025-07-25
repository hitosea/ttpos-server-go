<template>
  <template v-if="currency.unit_position == '0'">{{ currency.unit }}</template>
  <slot></slot>
  <template v-if="currency.unit_position == '1'">{{ currency.unit }}</template>
</template>
<script setup>
  import { useUserStore } from '@/store';
  const { currency } = useUserStore();
</script>
