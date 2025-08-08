<template>
  <div class="a-select">
    <el-select
      size="small"
      ref="selectRef"
      v-model="modelValue"
      :filterable="props.filterable"
      :placeholder="props.placeholder"
      @change="change"
    >
      <slot></slot>
    </el-select>
    <el-icon v-if="modelValue || modelValue === 0" class="close-icon" @click="clear"><CircleCloseFilled /></el-icon>
  </div>
</template>
<script setup>
  import { ref, computed } from 'vue';
  const props = defineProps({
    value: {
      type: [String, Number],
      default: null,
    },
    placeholder: {
      type: String,
      // 这里不要直接调用 $t，防止在组件外部环境未注入时报错
      default: '请选择',
    },
    filterable: {
      type: Boolean,
      default: false,
    },
  });
  const selectRef = ref(null);
  const emit = defineEmits(['change', 'update:value']);

  // 与父组件双向绑定，确保外部赋值时能同步到内部
  const modelValue = computed({
    get: () => props.value,
    set: (val) => {
      emit('update:value', val);
    },
  });

  const change = (val) => {
    // 只负责派发 change 事件，v-model 的更新由 modelValue 的 setter 完成
    emit('change', val);
  };
  const clear = () => {
    emit('update:value', null);
    emit('change', null);
  };
</script>
<style lang="scss" scoped>
  .a-select {
    position: relative;
    .close-icon {
      position: absolute;
      right: 8px;
      font-size: 18px;
      top: 8px;
      cursor: pointer;
      background: #fff;
    }
  }
</style>
