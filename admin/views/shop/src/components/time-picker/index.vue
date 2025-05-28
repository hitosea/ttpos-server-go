<template>
  <div class="time-picker max-w460">
    <el-time-picker class="have-icon" v-model="startTime" value-format="HH:mm" format="HH:mm" :clearable="false" />
    ~
    <el-time-picker v-model="endTime" value-format="HH:mm" format="HH:mm" :clearable="false" />
  </div>
</template>
<script setup>
  import { ref, watch, onMounted } from 'vue';
  const props = defineProps({
    modelValue: {
      type: Array,
      default: () => ['00:00', '23:59'],
    },
  });

  const emit = defineEmits(['update:modelValue']);

  // 使用默认值，避免 undefined
  const startTime = ref(props.modelValue[0] || '00:00');
  const endTime = ref(props.modelValue[1] || '23:59');

  // 添加 props 的监听，确保父组件更新值时能同步到子组件
  watch(
    () => props.modelValue,
    (newVal) => {
      if (newVal && newVal.length === 2) {
        startTime.value = newVal[0] || '00:00';
        endTime.value = newVal[1] || '23:59';
      }
    },
    { deep: true }
  );

  watch(startTime, (newVal) => {
    emit('update:modelValue', [newVal, endTime.value]);
  });
  watch(endTime, (newVal) => {
    emit('update:modelValue', [startTime.value, newVal]);
  });

  // 确保组件挂载时发送初始值
  onMounted(() => {
    if (!props.modelValue || props.modelValue.length !== 2) {
      emit('update:modelValue', [startTime.value, endTime.value]);
    }
  });
</script>
<style lang="scss" scoped>
  .time-picker {
    border: solid 1px var(--el-input-border-color, var(--el-border-color));
    border-radius: var(--el-input-border-radius, var(--el-border-radius-base));
  }
  :deep(.el-input__wrapper) {
    border: none !important;
    box-shadow: none !important;
    .el-input__inner {
      text-align: center;
    }
    .el-input__prefix {
      display: none;
    }
  }
  :deep(.have-icon) {
    .el-input__prefix {
      display: inline-flex;
    }
  }
</style>
