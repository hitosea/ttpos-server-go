<template>
  <div class="kitchen" v-loading="loading">
    <div class="form-div">
      <el-form size="small" ref="form" :model="form" label-position="top">
        <el-form-item for="no_click" :label="$t('厨显功能：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_open">
            <el-radio label="1">{{ $t('开') }}</el-radio>
            <el-radio label="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
          <div class="tips">{{ $t('关闭后将不在平板/扫码H5/点餐助手体现商品制作状态，且不将商品显示在厨显设备（不影响送厨打印）') }}</div>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('来菜提醒：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_come_dish">
            <el-radio label="1">{{ $t('开') }}</el-radio>
            <el-radio label="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('顾客呼叫提醒：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_call_service">
            <el-radio label="1">{{ $t('开') }}</el-radio>
            <el-radio label="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('等待时长颜色：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_wait_color">
            <el-radio label="1">{{ $t('开') }}</el-radio>
            <el-radio label="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item for="no_click" v-if="form.is_wait_color == 1" label="" :rules="[{ required: true, message: '' }]">
          <div class="max-w460 color-box">
            <el-input v-model="input1" disabled></el-input>
            <el-select v-model="form.wait_color[0]" size="default">
              <el-option value="red" :label="$t('红色')">{{ $t('红色') }}</el-option>
              <el-option value="yellow" :label="$t('黄色')">{{ $t('黄色') }}</el-option>
            </el-select>
          </div>
          <div class="max-w460 color-box">
            <el-input v-model="input2" disabled></el-input>
            <el-select v-model="form.wait_color[1]" size="default">
              <el-option value="red" :label="$t('红色')">{{ $t('红色') }}</el-option>
              <el-option value="yellow" :label="$t('黄色')">{{ $t('黄色') }}</el-option>
            </el-select>
          </div>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('高级设置密码')" prop="password" :rules="[{ required: true, message: '', trigger: 'chenge' }]">
          <el-input class="max-w460" v-model="password" type="password" disabled></el-input>
          <el-button @click="setPassword" type="primary" link size="small">{{ $t('设置密码') }}</el-button>
        </el-form-item>

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

        <div class="common-form">{{ $t('登录过的设备') }}</div>
        <el-form-item for="no_click" :label="$t('已登录的设备') + `(${onlineList.length}/${form.licenses?.kitchen_limit})`" class="form-items">
          <div class="form-div" v-for="(item, index) in onlineList" v-if="onlineList.length > 0">
            <div class="max-w460 input-ss">
              <autoTips :tooltipMaxWidth="460" :content="(item.remark ? item.remark : '') + `(${item.key})`">{{ (item.remark ? item.remark : '') + `(${item.key})` }}</autoTips>
            </div>
            <el-tooltip class="item" effect="dark" placement="top">
              <template #content>
                {{ $t('姓名') }}:{{ item.shopUser?.real_name || '-' }}<br />
                {{ $t('登录时间') }}:{{ item.finally_login_time || '-' }}<br />
                {{ $t('设备品牌') }}:{{ item.brand || '-' }}
              </template>
              <SvgIcon class="form-icon" name="man"></SvgIcon>
            </el-tooltip>
            <el-button @click="handleClick(item)" type="primary" link size="small">{{ $t('解绑') }}</el-button>
          </div>
          <p v-else>{{ $t('暂无设备') }}</p>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('已退出登录的设备')" class="form-items">
          <div class="form-div" v-for="(item, index) in offlineList" v-if="offlineList.length > 0">
            <div class="max-w460 input-ss">
              <autoTips :tooltipMaxWidth="460" :content="(item.remark ? item.remark : '') + `(${item.key})`">{{ (item.remark ? item.remark : '') + `(${item.key})` }}</autoTips>
            </div>
            <el-button @click="handleClick(item)" type="primary" link size="small">{{ $t('解绑') }}</el-button>
          </div>
          <p v-else>{{ $t('暂无设备') }}</p>
        </el-form-item>
      </el-form>
    </div>
    <div class="common-button-wrapper">
      <el-button size="small" @click="getData">{{ $t('重置') }}</el-button>
      <el-button size="small" type="primary" @click="onSubmit" :loading="loading">{{ $t('保存') }}</el-button>
    </div>
    <setPassword
      :advancedPassword="form.advanced_password"
      v-if="open"
      :open="open"
      @close="
        (e) => {
          open = false;
          if (e == 1) {
            this.getData();
          }
        }
      "
    >
    </setPassword>
  </div>
</template>
<script>
  import Terminal from '@/api/terminal.js';
  import IndexApi from '@/api/index.js';
  import setPassword from './setPassword.vue';
  import autoTips from '@/components/autoTips/autoTips.vue';
  import { useUserStore } from '@/store/index';
  import SvgIcon from '@/components/svg-icon/SvgIcon.vue';
  import { DTime } from '@/utils/DateTime.js';
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  export default {
    components: { setPassword, autoTips, SvgIcon },
    data() {
      return {
        supplier: supplier,
        form: {
          is_wait_color: 1,
          server: {
            ip: '',
            port: 8080,
          },
          language: [],
          default_language: null,
          advanced_password: false,
          wait_color: ['', ''],
          bind_list: [],
          is_come_dish: '1',
          is_call_service: '1',
          is_open: '1',
        },
        onlineList: [], //  在线设备列表
        offlineList: [], //离线设备列表
        origin: window.location.origin,
        port: window.location.port || '80',
        input1: $t('10分钟'),
        input2: $t('20分钟及以上'),
        languageList: [],
        open: false,
        loading: false,
        password: '',
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
      DTime: DTime,
      setPassword() {
        this.open = true;
      },
      getData() {
        let self = this;
        self.loading = true;
        Terminal.getKitchen()
          .then((data) => {
            self.loading = false;
            self.form = data.data.vars.values;
            self.languageList = data.data.vars.values.language_list;
            if (self.form.advanced_password) {
              self.password = 666666;
            }
            self.form.language = self.form.language.filter((lang) => {
              return self.languageList.map((h) => h.key).indexOf(lang) != -1;
            });
            self.onlineList = [];
            self.offlineList = [];
            self.form.bind_list.map((item) => {
              if (item.finally_login_id > 0) {
                self.onlineList.push(item);
              } else {
                self.offlineList.push(item);
              }
            });
          })
          .catch((error) => {
            self.loading = false;
          });
      },
      onSubmit() {
        let self = this;
        let params = JSON.parse(JSON.stringify(self.form));
        //绑定的设备不用提清空
        params.bind_list = [];
        self.loading = true;
        Terminal.saveKitchen(params, true)
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

      //解绑
      handleClick(item) {
        let self = this;
        ElMessageBox.confirm($t('解绑后不可恢复，确认解绑吗?'), $t('提示'), {
          confirmButtonText: $t('确定'),
          cancelButtonText: $t('取消'),
          type: 'warning',
        })
          .then(() => {
            self.loading = true;
            let params = {
              bind_id: item.id,
            };
            IndexApi.unbind(params, true)
              .then((data) => {
                self.loading = false;
                this.$ElMessage({
                  message: $t('保存成功'),
                  type: 'success',
                });
                self.getData();
              })
              .catch((error) => {
                self.loading = false;
              });
          })
          .catch(() => {});
      },
    },
  };
</script>
<style scoped lang="scss">
  .color-box {
    display: flex;
    gap: 12px;
    margin-right: 16px;
  }

  .el-button--primary:focus {
    color: white;
  }

  .kitchen {
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
        .form-icon {
          color: var(--el-color-primary);
          width: 30px;
          height: 30px;
          margin-left: 8px;
        }
      }
    }
  }
</style>
