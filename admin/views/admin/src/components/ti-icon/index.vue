<template>
  <svg class="svg-icon" aria-hidden="true" :style="styles">
    <use :xlink:href="symbolId" fill="currentColor" />
  </svg>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import type { CSSProperties } from 'vue';

  const props = withDefaults(
    defineProps<{
      name: string;
      prefix?: string;
      size?: number | string;
      color?: string;
    }>(),
    {
      prefix: 'icon',
      size: 16,
      color: 'inherit',
    },
  );

  const symbolId = computed(() => `#${props.prefix}-${props.name}`);
  const styles = computed<CSSProperties>(() => {
    return {
      width: addUnit(props.size),
      height: addUnit(props.size),
      color: props.color,
    };
  });

  const addUnit = (value: string | number, unit = 'px') => {
    return !Object.is(Number(value), NaN) ? `${value}${unit}` : value;
  };
</script>

<style lang="scss" scoped>
  .svg-icon {
    vertical-align: -0.1em;
    fill: currentColor;
    overflow: hidden;
  }
</style>
