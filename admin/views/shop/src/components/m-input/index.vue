<template>
  <div class="m-input" :id="id + '-input'" @click="handleFocus">
    <el-input v-model="valueData" :disabled="disabled" :placeholder="placeholder" @input="handleInput" :class="width" :maxlength="maxlength"> </el-input>
    <el-button :id="id + '-button'" style="display: none" size="small" type="primary" :disabled="buttonDisabled" @click="onSubmit" :loading="loading">{{
      $t('点击翻译')
    }}</el-button>
  </div>
</template>
<script>
  import IndexApi from '@/api/index.js';
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
      disabled: {
        type: Boolean,
        default: false,
      },
      width: {
        type: String,
        default: 'max-w460',
      },
      langKey: {
        type: String,
        default: '',
      },
      maxlength: {
        type: Number,
        default: 255,
      },
    },
    data() {
      return {
        loading: false,
        valueData: '',
        id: this.randomString(16),
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
    computed: {
      buttonDisabled() {
        return this.value == '';
      },
    },
    mounted() {
      let self = this;
      let myDiv = document.getElementById(this.id + '-input');
      // 添加事件监听器到document对象
      document.addEventListener('click', function (event) {
        // 检查点击的目标元素是否是myDiv或其子元素
        if (myDiv && !myDiv.contains(event.target)) {
          // 如果点击的是myDiv以外的元素，执行相应的处理程序
          self.handleBlur();
        }
      });
    },
    methods: {
      randomString(len) {
        const $chars = 'ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678oOLl9gqVvUuI1';
        const maxPos = $chars.length;
        let pwd = '';
        for (let i = 0; i < len; i++) {
          pwd += $chars.charAt(Math.floor(Math.random() * maxPos));
        }
        return pwd;
      },

      onSubmit() {
        this.getLang();
      },
      getLang() {
        let data = { data: [] };
        data.data.push({
          lang: this.langKey,
          content: this.value,
        });
        this.loading = true;
        IndexApi.aiTranslate(data, true)
          .then((res) => {
            const jsonString = res.data.data;
            const array = JSON.parse(jsonString);
            this.$emit('translate', array);
            this.handleBlur();
            this.loading = false;
          })
          .catch((error) => {
            this.loading = false;
          });
      },

      handleInput($event) {
        this.$emit('update:valueData', $event);
        this.$emit('change', $event);
        this.$nextTick(() => {
          let myButton = document.getElementById(this.id + '-button');
          if (myButton) {
            myButton.style.display = 'block';
          }
        });
      },

      handleFocus() {
        this.$nextTick(() => {
          let myButton = document.getElementById(this.id + '-button');
          if (myButton) {
            myButton.style.display = 'block';
          }
        });
      },

      handleBlur() {
        let myButton = document.getElementById(this.id + '-button');
        if (myButton) {
          myButton.style.display = 'none';
        }
      },
    },
  };
</script>
<style lang="scss" scoped>
  .m-input {
    width: 100%;
    display: flex;
    gap: 12px;
  }
  .m-full {
    width: 100%;
  }
  .w-100px {
    width: 100px;
  }
</style>
