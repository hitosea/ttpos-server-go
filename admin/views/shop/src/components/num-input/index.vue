<template>
  <div :class="width" class="flex-num-input">
    <div v-if="controls" class="icon-plus" @click="handlePlus">
      <el-icon><Plus /></el-icon>
    </div>
    <el-input
      :class="controls ? 'controls-input' : ''"
      :disabled="disabled"
      :model-value="modelValue"
      type="text"
      :placeholder="placeholder"
      @input="handleInput"
      @blur="handleBlur"
    >
    </el-input>
    <div v-if="controls" class="icon-minus" @click="handleMinus">
      <el-icon><Minus /></el-icon>
    </div>
  </div>
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
      controls: {
        type: Boolean,
        default: false,
      },
      disabled: {
        type: Boolean,
        default: false,
      },
    },
    methods: {
      handleMinus() {
        if (this.disabled) {
          return;
        }
        let newValue = Number(this.modelValue) - 1;
        // Round to the specified precision to avoid floating point issues
        if (this.precision > 0) {
          const factor = Math.pow(10, this.precision);
          newValue = Math.round(newValue * factor) / factor;
        }
        // Ensure the value doesn't go below min
        newValue = Math.max(newValue, this.min);
        this.$emit('update:modelValue', newValue);
      },
      handlePlus() {
        if (this.disabled) {
          return;
        }
        let newValue = Number(this.modelValue) + 1;
        // Round to the specified precision to avoid floating point issues
        if (this.precision > 0) {
          const factor = Math.pow(10, this.precision);
          newValue = Math.round(newValue * factor) / factor;
        }
        // Ensure the value doesn't exceed max
        newValue = Math.min(newValue, this.max);
        this.$emit('update:modelValue', newValue);
      },

      async handleBlur() {
        await this.handleChange(this.modelValue);
        // 处理 类似2.20 去掉末尾的0, 2.00 去掉0 变成2；2.0 变成2
        this.$emit('update:modelValue', this.removeTrailingZeros(this.modelValue));
      },

      handleInput(value) {
        this.$emit('update:modelValue', value);
        // 如果最小值包含小数点，则不进行处理
        if (this.min.toString().includes('.')) {
          return;
        }
        this.handleChange(value);
      },

      handleChange(value) {
        if (this.disabled || !value) {
          return;
        }
        // 只允许数字和小数点
        let formattedValue = value.toString()
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

      // 移除末尾的0
      removeTrailingZeros(value) {
        if (!value) {
          return value;
        }
        if (value.includes('.')) {
          const [integer, decimal] = value.split('.');
          if (decimal) {
            const cleanDecimal = decimal.replace(/0+$/, '');
            if (cleanDecimal === '') {
              return integer; // 如果小数部分全是0，返回整数部分
            }
            return `${integer}.${cleanDecimal}`;
          }
        }
        return value;
      },
    },
  };
</script>

<style scoped lang="scss">
  .m-full {
    width: 100%;
  }
  .flex-num-input {
    display: flex;
    flex-direction: row;
    align-items: center;
    width: 100%;
  }
  .icon-plus {
    cursor: pointer;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #f5f5f5;
    border: 1px solid #ccc;
    border-right: none;
    border-radius: 4px 0 0 4px;
  }
  .icon-minus {
    cursor: pointer;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #f5f5f5;
    border: 1px solid #ccc;
    border-left: none;
    border-radius: 0 4px 4px 0;
  }
  .controls-input {
    border-radius: 0 !important;
    :deep(.el-input__wrapper) {
      border-radius: 0 !important;
    }
  }
</style>
