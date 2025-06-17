<template>
  <el-input v-model="valueData" type="text" :placeholder="placeholder" @input="changeTax" :class="width"> </el-input>
</template>
<script>
  export default {
    props: {
      value: {
        type: String,
        required: true,
        default: '',
      },
      placeholder: {
        type: String,
        default: $t('请输入'),
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
    data() {
      return {
        valueData: '',
      };
    },
    watch: {
      value: {
        handler(newVal) {
          this.valueData = newVal;
        },
        deep: true,
        immediate: true,
      },
    },

    methods: {
      changeTax() {
        var regex = /^[0-9]+(\.[0-9]+)?$/;
        if (!regex.test(this.valueData)) {
          this.valueData = this.valueData
            .replace(/[^0-9.]/g, '')
            .replace(/\.{2,}/g, '.')
            .replace(/(\..*)\./g, '$1');
        }
        // 将数字转换为字符串
        let str = this.valueData.toString();
        // 检查字符串是否包含小数点
        if (str.includes('.')) {
          // 获取小数点后的部分
          let decimalPart = str.split('.')[1];
          // 检查小数点后的部分是否大于两位
          if (decimalPart.length > this.precision) {
            // 获取整数部分和小数部分
            let parts = str.split('.');
            // 直接截取需要的小数位数，不进行四舍五入
            this.valueData = parts[0] + '.' + parts[1].substring(0, this.precision);
          }
        }
        this.valueData > this.max ? (this.valueData = this.max) : (this.valueData = this.valueData);
        this.valueData < this.min ? (this.valueData = this.min) : (this.valueData = this.valueData);
        this.$emit('update:valueData', this.valueData);
      },
    },
  };
</script>
<style scoped>
  .m-full {
    width: 100%;
  }
</style>
