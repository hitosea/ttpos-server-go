<template>
  <div class="assistant" v-loading="loading">
    <div class="form-div">
      <el-form size="small" ref="form" :model="form" label-position="top">
        <el-form-item for="no_click" :label="$t('自助餐剩余时长颜色：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_remain_color">
            <el-radio value="1">{{ $t('开') }}</el-radio>
            <el-radio value="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item for="no_click" v-if="form.is_remain_color == 1" label="" :rules="[{ required: true, message: '' }]">
          <div class="max-w460 color-box">
            <el-input v-model="input1" disabled></el-input>
            <el-select v-model="form.remain_color[0]" size="default">
              <el-option value="#E50028" :label="$t('红色')">{{ $t('红色') }}</el-option>
              <el-option value="#F2A000" :label="$t('黄色')">{{ $t('黄色') }}</el-option>
            </el-select>
          </div>
          <div class="max-w460 color-box">
            <el-input v-model="input2" disabled></el-input>
            <el-select v-model="form.remain_color[1]" size="default">
              <el-option value="#E50028" :label="$t('红色')">{{ $t('红色') }}</el-option>
              <el-option value="#F2A000" :label="$t('黄色')">{{ $t('黄色') }}</el-option>
            </el-select>
          </div>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('高级设置密码')" prop="password" :rules="[{ required: true, message: '', trigger: 'chenge' }]">
          <el-input class="max-w460" v-model="password" type="password" disabled></el-input>
          <el-button @click="setPassword($t('高级设置密码'))" type="primary" link size="small">{{ $t('设置密码') }}</el-button>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('下单校验高级密码')">
          <el-radio-group v-model="form.is_check_order">
            <el-radio value="1">{{ $t('开') }}</el-radio>
            <el-radio value="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('锁屏密码')" prop="password" :rules="[{ required: true, message: '', trigger: 'chenge' }]">
          <el-input class="max-w460" v-model="lockPassword" type="password" disabled></el-input>
          <el-button @click="setPassword($t('锁屏密码'))" type="primary" link size="small">{{ $t('设置密码') }}</el-button>
        </el-form-item>

        <!-- <el-form-item for="no_click" :label="$t('支持功能')">
          <el-checkbox-group v-model="form.support_function">
            <template v-for="item in support_function_list">
              <el-checkbox :label="item.key">{{ $t(item.name) }}</el-checkbox>
            </template>
          </el-checkbox-group>
          <div class="tips">{{ $t('注：勾选后在服务员模式中可操作对应功能') }}</div>
        </el-form-item> -->

        <el-form-item for="no_click" :label="$t('自动锁屏')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_auto_lock_screen">
            <el-radio value="1">{{ $t('开') }}</el-radio>
            <el-radio value="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item for="no_click" v-if="form.is_auto_lock_screen == '1'" label="" prop="auto_lock_screen" :rules="[{ required: true, message: '' }]">
          <el-select v-model="form.auto_lock_screen">
            <template v-for="(item, index) in lockList" :key="index">
              <el-option :value="item.key" :label="item.label">{{ item.label }}</el-option>
            </template>
          </el-select>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('常用语言')" prop="language" :rules="[{ required: true, message: $t('请选择常用语言') }]">
          <el-checkbox-group v-model="form.language">
            <el-checkbox v-for="item in languageList" v-show="item.key" :key="item.key" :value="item.key" :disabled="form.language.length == 1 && form.language.includes(item.key)">
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
        <el-form-item for="no_click" :label="$t('已登录的设备') + `(${onlineList.length}/${form.licenses?.assistant_limit})`" class="form-items">
          <div class="form-div" v-for="(item, index) in onlineList" v-if="onlineList.length > 0">
            <div class="max-w460 input-ss">
              <autoTips :tooltipMaxWidth="460" :content="(item.remark ? item.remark : '') + `(${item.key})`">{{ (item.remark ? item.remark : '') + `(${item.key})` }}</autoTips>
            </div>
            <el-tooltip class="item" effect="dark" placement="top">
              <template #content>
                {{ $t('姓名') }}:{{ item.shopUser?.real_name || '-' }}<br />
                {{ $t('登录时间') }}:{{ item.finally_login_time || '-' }}<br />
                {{ $t('设备品牌') }}:{{ item.brand || '-' }}<br />
                <template v-if="item.version"> {{ $t('版本号') }}:{{ item.version }}</template>
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
      :openTitle="openTitle"
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
  import SvgIcon from '@/components/svg-icon/SvgIcon.vue';
  import { useUserStore } from '@/store/index';
  import { DTime } from '@/utils/DateTime.js';
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  export default {
    components: { setPassword, autoTips, SvgIcon },
    data() {
      return {
        supplier: supplier,
        form: {
          default_mode: 1,
          server: {
            ip: '',
            port: 8080,
          },
          language: [],
          default_language: null,
          advanced_password: false,
          lock_password: false,
          is_remain_color: '1',
          remain_color: ['', ''],
          bind_list: [],
          support_function: [],
          is_check_order: '0',
        },
        onlineList: [], //  在线设备列表
        offlineList: [], //离线设备列表
        support_function_list: [],
        origin: window.location.origin,
        port: window.location.port || '80',
        input1: $t('10分钟以内'),
        input2: $t('20分钟以内'),
        languageList: [],
        lockList: [
          {
            label: $t('无操作15秒'),
            key: '15',
          },
          {
            label: $t('无操作30秒'),
            key: '30',
          },
          {
            label: $t('无操作1分钟'),
            key: '60',
          },
          {
            label: $t('无操作2分钟'),
            key: '120',
          },
          {
            label: $t('无操作5分钟'),
            key: '300',
          },
          {
            label: $t('无操作10分钟'),
            key: '600',
          },
        ],
        open: false,
        openTitle: '',
        loading: false,
        password: '',
        lockPassword: '',
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
      setPassword(e) {
        this.open = true;
        this.openTitle = e;
      },
      getData() {
        let self = this;
        self.loading = true;
        Terminal.getAssistant()
          .then((data) => {
            self.loading = false;
            self.form = data.data.vars.values;
            self.languageList = data.data.vars.values.language_list;
            self.support_function_list = data.data.vars.values.support_function_list;
            if (self.form.advanced_password) {
              self.password = 666666;
            }
            if (self.form.lock_password) {
              self.lockPassword = 666666;
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
        Terminal.saveAssistant(params, true)
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

  .assistant {
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
