<template>
  <template v-if="currency.is_open == 1">
    <template v-if="currency.vices.vice_unit_position == '0'">{{ currency.vices?.vice_unit }}</template>
    <slot></slot>
    <template v-if="currency.vices.vice_unit_position == '1'">{{ currency.vices?.vice_unit }} </template>
  </template>
</template>
<script setup>
  import { useUserStore } from '@/store';
  const { currency } = useUserStore();
</script>
