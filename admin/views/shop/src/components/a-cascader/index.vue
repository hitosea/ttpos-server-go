<template>
  <div class="a-cascader">
    <el-cascader ref="cascaderRef" :options="props.options" :props="props.props" v-model="value" :placeholder="props.placeholder" @change="change">
      <template #default="{ data }">
        <span class="span-w" @click="handleValue(data)">{{ data.label }}</span>
      </template>
    </el-cascader>
    <el-icon v-if="value" class="close-icon" @click="clear"><CircleCloseFilled /></el-icon>
  </div>
</template>
<script setup>
  import { ref } from 'vue';
  const props = defineProps({
    options: {
      type: Array,
      default: [],
    },
    placeholder: {
      type: String,
      default: $t('请选择'),
    },
    props: {
      type: Object,
      default: {},
    },
    value: {
      type: Array,
      default: [],
    },
  });
  const emit = defineEmits(['update:value', 'change']);
  const value = ref(props.value);
  const cascaderRef = ref(null);
  const change = (data) => {
    value.value = data;
    cascaderRef.value.togglePopperVisible();
    emit('update:value', data);
    emit('change', data);
  };
  const handleValue = (data) => {
    value.value = data.value;
    cascaderRef.value.togglePopperVisible();
    emit('update:value', data.value);
    emit('change', data.value);
  };
  const clear = () => {
    value.value = '';
    emit('update:value', '');
    emit('change', '');
  };
</script>
<style lang="scss" scoped>
  .a-cascader {
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
  .span-w {
    display: block;
    width: 100%;
  }
</style>
