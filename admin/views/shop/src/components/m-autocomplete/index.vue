<template>
  <div class="m-autocomplete" :id="id + '-autocomplete'" @click="handleFocus">
    <el-autocomplete
      :fetch-suggestions="(e, h) => querySearch(e, h, numKey)"
      @select="(e) => $emit('select', e)"
      v-model="valueData"
      :placeholder="placeholder"
      :class="width"
      :maxlength="maxlength"
    >
    </el-autocomplete>
    <el-button :id="id + '-button'" style="display: none" size="small" type="primary" @click="onSubmit" :loading="loading">{{ $t('点击翻译') }}</el-button>
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
      width: {
        type: String,
        default: 'max-w460',
      },
      //提交的数据key 0,1,2,3
      numKey: {
        type: [String, Number],
        default: '',
      },
      //语言的key zh,en,ja等等....
      langKey: {
        type: String,
        default: '',
      },
      maxlength: {
        type: Number,
        default: 255,
      },
      restaurantsObj: {
        type: [Array, Object],
        default: () => null,
      },
    },
    data() {
      return {
        loading: false,
        valueData: '',
        id: this.randomString(18),
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
      valueData: {
        handler(newVal) {
          this.$emit('update:valueData', newVal);
        },
        deep: true,
        immediate: true,
      },
    },
    mounted() {
      let self = this;
      let myDiv = document.getElementById(this.id + '-autocomplete');
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
        if (!this.value) return;
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

      querySearch(queryString, cb, key) {
        let restaurants = [];
        restaurants = this.restaurantsObj[key];
        let results = queryString ? restaurants.filter(this.createFilter(queryString, key)) : restaurants;
        // 调用 callback 返回建议列表的数据
        cb(results);
      },

      createFilter(queryString, key) {
        let restaurants = [];
        restaurants = this.restaurantsObj[key];
        return (restaurants) => {
          return restaurants.value.toLowerCase().indexOf(queryString.toLowerCase()) === 0;
        };
      },
    },
  };
</script>
<style scoped>
  .m-autocomplete {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 12px;
  }
  :deep(.el-autocomplete) {
    width: 100%;
  }
  .w-full {
    width: 100%;
  }
</style>
