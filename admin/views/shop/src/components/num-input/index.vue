<template>
  <el-input :model-value="modelValue" type="text" :placeholder="placeholder" @input="handleInput" :class="width"> </el-input>
</template>

<script>
  export default {
    name: 'NumInput',
    props: {
      modelValue: {
        type: [String, Number],
        default: '',
      },
      placeholder: {
        type: String,
        default: () => $t('请输入'),
      },
      width: {
        type: String,
        default: 'max-w460',
      },
      min: {
        type: Number,
        default: -Infinity,
      },
      max: {
        type: Number,
        default: Infinity,
      },
      precision: {
        type: Number,
        default: 4,
      },
    },
    methods: {
      handleInput(value) {
        // 只允许数字和小数点
        let formattedValue = value
          .replace(/[^0-9.]/g, '')
          .replace(/\.{2,}/g, '.')
          .replace(/(\..*)\./g, '$1');

        // 处理以0开头的数字
        if (formattedValue.length > 1 && formattedValue.startsWith('0') && !formattedValue.startsWith('0.')) {
          formattedValue = formattedValue.slice(1);
        }

        //如果precision为0，
        if (this.precision === 0) {
          formattedValue = formattedValue.replace(/\./g, '');
        }

        // 处理小数位数
        if (formattedValue.includes('.')) {
          const [integer, decimal] = formattedValue.split('.');
          if (decimal && decimal.length > this.precision) {
            formattedValue = `${integer}.${decimal.slice(0, this.precision)}`;
          }
        }

        // 处理最小最大值
        const numValue = parseFloat(formattedValue);
        if (!isNaN(numValue)) {
          if (numValue > this.max) {
            formattedValue = this.max.toString();
          } else if (numValue < this.min) {
            formattedValue = this.min.toString();
          }
        }

        // 触发更新
        this.$emit('update:modelValue', formattedValue);
      },
    },
  };
</script>

<style scoped>
  .m-full {
    width: 100%;
  }
</style>
