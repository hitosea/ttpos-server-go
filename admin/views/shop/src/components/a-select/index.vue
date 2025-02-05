<template>
  <div class="a-select">
    <el-select size="small" ref="selectRef" v-model="value" :filterable="props.filterable" :placeholder="props.placeholder" @change="change">
      <slot></slot>
    </el-select>
    <el-icon v-if="value || value === 0" class="close-icon" @click="clear"><CircleCloseFilled /></el-icon>
  </div>
</template>
<script setup>
  import { ref } from 'vue';
  const props = defineProps({
    value: {
      type: String || Number,
      default: null,
    },
    placeholder: {
      type: String,
      default: $t('请选择'),
    },
    filterable: {
      type: Boolean,
      default: false,
    },
  });
  const selectRef = ref(null);
  const value = ref(props.value);
  const emit = defineEmits(['change', 'update:value']);
  const change = (value) => {
    emit('update:value', value);
    emit('change', value);
  };
  const clear = () => {
    value.value = null;
    emit('update:value', null);
    emit('change', value);
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
