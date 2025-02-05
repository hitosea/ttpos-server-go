<template>
  <div class="salb">
    <div class="form-div" v-loading="loading">
      <el-form size="small" ref="form" :model="form" label-position="top">
        <el-form-item for="no_click" :label="$t('呼叫服务员：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_call_service">
            <el-radio label="1">{{ $t('开') }}</el-radio>
            <el-radio label="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('顾客可开桌：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_customer_order">
            <el-radio label="1">{{ $t('开') }}</el-radio>
            <el-radio label="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('声音提醒：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_voice_remind">
            <el-radio label="1">{{ $t('开') }}</el-radio>
            <el-radio label="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item for="no_click" v-if="is_open_buffet" :label="$t('自助餐下单限制：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_buffet_order_limit" @change="handleChangeBuff">
            <el-radio label="1">{{ $t('开') }}</el-radio>
            <el-radio label="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <template v-if="form.is_buffet_order_limit == '1'">
          <el-form-item for="no_click" label="" :rules="[{ required: true, message: '' }]">
            <el-checkbox v-model="form.buffet_order_limit.is_limit_time" :disabled="form.buffet_order_limit.is_limit_num == '0'" true-label="1" false-label="0">{{
              $t('时间限制')
            }}</el-checkbox>
            <template v-if="form.buffet_order_limit.is_limit_time == '1'">
              <el-input-number
                :controls="false"
                :precision="0"
                :min="1"
                :max="999"
                style="width: 200px !important; margin: 0 8px"
                :placeholder="$t('请输入时间')"
                v-model.number="form.buffet_order_limit.limit_time"
              ></el-input-number>
              {{ $t('分') }}
            </template>
            <div class="gray9">{{ $t('设置下单后间隔多久才可再次下单') }}</div>
          </el-form-item>

          <el-form-item for="no_click" label="" :rules="[{ required: true, message: '' }]">
            <el-checkbox v-model="form.buffet_order_limit.is_limit_num" :disabled="form.buffet_order_limit.is_limit_time == '0'" true-label="1" false-label="0">{{
              $t('数量限制')
            }}</el-checkbox>
            <template v-if="form.buffet_order_limit.is_limit_num == '1'">
              <el-input-number
                :controls="false"
                :precision="0"
                :min="1"
                :max="999"
                style="width: 200px !important; margin: 0 8px"
                :placeholder="$t('请输入数量')"
                v-model.number="form.buffet_order_limit.limit_num"
              ></el-input-number>
            </template>
            <div class="gray9">{{ $t('设置每次下单的最大商品总数') }}</div>
          </el-form-item>
        </template>

        <el-form-item for="no_click" :label="is_open_buffet ? $t('非自助餐下单限制：') : $t('下单限制：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_order_limit" @change="handleChangeOrder">
            <el-radio label="1">{{ $t('开') }}</el-radio>
            <el-radio label="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <template v-if="form.is_order_limit == '1'">
          <el-form-item for="no_click" label="" :rules="[{ required: true, message: '' }]">
            <el-checkbox v-model="form.order_limit.is_limit_time" :disabled="form.order_limit.is_limit_num == '0'" true-label="1" false-label="0">{{ $t('时间限制') }}</el-checkbox>
            <template v-if="form.order_limit.is_limit_time == '1'">
              <el-input-number
                :controls="false"
                :precision="0"
                :min="1"
                :max="999"
                style="width: 200px !important; margin: 0 8px"
                :placeholder="$t('请输入时间')"
                v-model.number="form.order_limit.limit_time"
              ></el-input-number>
              {{ $t('分') }}
            </template>
            <div class="gray9">{{ $t('设置下单后间隔多久才可再次下单') }}</div>
          </el-form-item>

          <el-form-item for="no_click" label="" :rules="[{ required: true, message: '' }]">
            <el-checkbox v-model="form.order_limit.is_limit_num" :disabled="form.order_limit.is_limit_time == '0'" true-label="1" false-label="0">{{ $t('数量限制') }}</el-checkbox>
            <template v-if="form.order_limit.is_limit_num == '1'">
              <el-input-number
                :controls="false"
                :precision="0"
                :min="1"
                :max="999"
                style="width: 200px !important; margin: 0 8px"
                :placeholder="$t('请输入数量')"
                v-model.number="form.order_limit.limit_num"
              ></el-input-number>
            </template>
            <div class="gray9">{{ $t('设置每次下单的最大商品总数') }}</div>
          </el-form-item>
        </template>

        <el-form-item for="no_click" :label="$t('常用语言')" prop="language" :rules="[{ required: true, message: $t('请选择常用语言') }]">
          <el-checkbox-group v-model="form.language">
            <el-checkbox v-for="item in languageList" v-show="item.key" :key="item.key" :label="item.key" :disabled="form.language.length == 1 && form.language.includes(item.key)">
              {{ item.value }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('默认语言')" prop="default_language" :rules="[{ required: true, message: $t('请选择常用语言') }]">
          <el-select v-model="form.default_language">
            <template v-for="(item, index) in defaultLanguageList" :key="index">
              <el-option :value="item.key" :label="item.value">{{ item.value }}</el-option>
            </template>
          </el-select>
        </el-form-item>
      </el-form>
    </div>
    <div class="common-button-wrapper">
      <el-button size="small" @click="getData()">{{ $t('重置') }}</el-button>
      <el-button size="small" type="primary" @click="onSubmit" :loading="loading">{{ $t('保存') }}</el-button>
    </div>
  </div>
</template>
<script>
  import Terminal from '@/api/terminal.js';
  import { useUserStore } from '@/store/index';
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const is_open_buffet = supplier.value?.is_open_buffet || 0;
  export default {
    data() {
      return {
        is_open_buffet: is_open_buffet,
        supplier: supplier,
        form: {
          is_call_service: '1',
          is_customer_order: '1',
          is_voice_remind: '1',
          is_buffet_order_limit: '1',
          buffet_order_limit: {
            is_limit_time: '1',
            limit_time: null,
            is_limit_num: '1',
            limit_num: null,
          },
          is_order_limit: '1',
          order_limit: {
            is_limit_time: '1',
            limit_time: null,
            is_limit_num: '1',
            limit_num: null,
          },
          language: [],
          default_language: null,
        },
        languageList: [],
        loading: false,
      };
    },
    created() {
      this.getData();
      var IPRegex = /^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$/;
      if (IPRegex.test(window.location.hostname)) {
        this.origin = window.location.hostname;
      } else {
        this.origin = window.location.protocol + '//' + window.location.hostname;
      }
    },
    computed: {
      defaultLanguageList() {
        let result = [];
        this.languageList.map((item) => {
          if ((this.form.language || []).includes(item.key)) {
            result.push(item);
          }
          if (!(this.form.language || []).includes(this.form.default_language)) {
            this.form.default_language = this.form.language[0];
          }
        });
        return result;
      },
    },
    methods: {
      onSubmit() {
        let self = this;

        setTimeout(() => {
          const errorItems = document.querySelectorAll('.el-form-item__error');
          console.log(errorItems);
          if (errorItems.length > 0) {
            const firstErrorItem = errorItems[0];
            firstErrorItem.scrollIntoView({ behavior: 'smooth', block: 'center' });
          }
        }, 200);

        let params = JSON.parse(JSON.stringify(self.form));
        self.loading = true;
        Terminal.saveH5(params, true)
          .then((data) => {
            self.loading = false;
            this.$ElMessage({
              message: $t('保存成功'),
              type: 'success',
            });
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      getData() {
        let self = this;
        self.loading = true;
        Terminal.getH5()
          .then((data) => {
            self.loading = false;
            self.form = data.data.vars.values;
            self.languageList = data.data.vars.values.language_list;

            self.form.language = self.form.language.filter((lang) => {
              return self.languageList.map((h) => h.key).indexOf(lang) != -1;
            });
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      handleChangeBuff(e) {
        if (e == '1') {
          this.form.buffet_order_limit.is_limit_time = '1';
          this.form.buffet_order_limit.is_limit_num = '1';
        }
      },

      handleChangeOrder(e) {
        if (e == '1') {
          this.form.order_limit.is_limit_time = '1';
          this.form.order_limit.is_limit_num = '1';
        }
      },
    },
  };
</script>
<style scoped lang="scss">
  .salb {
    height: calc(100% - 14px);
    display: flex;
    flex-direction: column;
    overflow: hidden;

    .form-div {
      flex: 1 1 auto;
      overflow-y: auto;
      overflow-x: hidden;
    }

    .common-button-wrapper {
      flex: 0 0 auto;
      flex-shrink: 0;
    }
  }

  .delete-box {
    display: flex;
    align-items: center;
    cursor: pointer;
  }

  .delete-box-one {
    color: var(--el-color-tips);
    cursor: not-allowed;
  }

  .el-button--primary:focus {
    color: white;
  }

  .form-items {
    :deep(.el-form-item__content) {
      display: flex;
      flex-direction: column;
      align-items: start;
      gap: 12px;

      .form-div {
        display: flex;
        width: 100%;

        .input-ss {
          background-color: var(--el-disabled-bg-color);
          box-shadow: 0 0 0 1px var(--el-disabled-border-color) inset;
          background-image: none;
          border-radius: var(--el-input-border-radius, var(--el-border-radius-base));
          cursor: text;
          transition: var(--el-transition-box-shadow);
          transform: translate3d(0, 0, 0);
          width: 100%;
          position: relative;
          padding: 0;
        }
      }
    }
  }
</style>
