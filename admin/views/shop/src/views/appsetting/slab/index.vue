<template>
  <div class="salb" v-loading="loading">
    <div class="form-div">
      <el-form size="small" ref="form" :model="form" label-position="top">
        <el-form-item for="no_click" :label="$t('轮播内容')" prop="carousel" :rules="[{ required: true, message: '' }]">
          <div class="draggable-list">
            <flieUpload
              @upLoad="upLoad"
              :imgSize="2"
              :videoSize="10"
              :source="'tablet'"
              :tips="[$t('图片：支持JPG、JPEG、PNG、WEBP格式，小于2MB，尺寸：1160*1104px'), $t('视频：支持MP4格式，小于10MB，尺寸：1160*1104px')]"
            ></flieUpload>
            <el-table size="small" :data="form.carousel" border style="width: 100%" v-loading="loading">
              <el-table-column prop="real_name" :label="$t('图片名称')"></el-table-column>
              <el-table-column prop="sort" :label="$t('排序')">
                <template #default="scope">
                  <el-form-item
                    for="no_click"
                    ref="form-item"
                    style="margin-top: 16px"
                    :rules="[
                      {
                        required: true,
                        validator: () => {
                          return scope.row.sort >= 0 && typeof scope.row.sort == 'number' ? true : false;
                        },
                        message: $t('请输入排序'),
                      },
                    ]"
                    prop="scope.row.sort"
                  >
                    <el-input-number :controls="false" :min="0" :max="999" :precision="0" :placeholder="$t('接近0，排序等级越高')" v-model="scope.row.sort"></el-input-number>
                  </el-form-item>
                </template>
              </el-table-column>
              <el-table-column prop="file_path" :label="$t('链接地址')">
                <template #default="scope">
                  <el-input v-model="scope.row.file_path" disabled></el-input>
                </template>
              </el-table-column>
              <el-table-column prop="file_path" width="100" :label="$t('操作')">
                <template #default="scope">
                  <div class="delete-box" :class="form.carousel.length == 1 ? 'delete-box-one' : ''">
                    <el-icon size="24">
                      <Delete @click="deleteOne(scope)" />
                    </el-icon>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('呼叫服务员：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_call_service">
            <el-radio value="1">{{ $t('开') }}</el-radio>
            <el-radio value="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('顾客可开桌：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_customer_order">
            <el-radio value="1">{{ $t('开') }}</el-radio>
            <el-radio value="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('声音提醒：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_voice_remind">
            <el-radio value="1">{{ $t('开') }}</el-radio>
            <el-radio value="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item for="no_click" v-if="is_open_buffet" :label="$t('自助餐下单限制：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_buffet_order_limit" @change="handleChangeBuff">
            <el-radio value="1">{{ $t('开') }}</el-radio>
            <el-radio value="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <template v-if="form.is_buffet_order_limit == '1'">
          <el-form-item for="no_click" label="" :rules="[{ required: true, message: '' }]">
            <el-checkbox v-model="form.buffet_order_limit.is_limit_time" :disabled="form.buffet_order_limit.is_limit_num == '0'" true-value="1" false-value="0">{{
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
            <el-checkbox v-model="form.buffet_order_limit.is_limit_num" :disabled="form.buffet_order_limit.is_limit_time == '0'" true-value="1" false-value="0">{{
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
            <el-radio value="1">{{ $t('开') }}</el-radio>
            <el-radio value="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <template v-if="form.is_order_limit == '1'">
          <el-form-item for="no_click" label="" :rules="[{ required: true, message: '' }]">
            <el-checkbox v-model="form.order_limit.is_limit_time" :disabled="form.order_limit.is_limit_num == '0'" true-value="1" false-value="0">{{ $t('时间限制') }}</el-checkbox>
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
            <el-checkbox v-model="form.order_limit.is_limit_num" :disabled="form.order_limit.is_limit_time == '0'" true-value="1" false-value="0">{{ $t('数量限制') }}</el-checkbox>
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

        <el-form-item for="no_click" :label="$t('高级设置密码')" prop="password" :rules="[{ required: true, message: '', trigger: 'chenge' }]">
          <el-input class="max-w460" v-model="password" type="password" disabled></el-input>
          <el-button @click="setPassword" type="primary" link size="small">{{ $t('设置密码') }}</el-button>
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
        <el-form-item for="no_click" :label="$t('已登录的设备') + `(${onlineList.length}/${form.licenses?.tablet_limit})`" class="form-items">
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
      <el-button size="small" @click="getData()">{{ $t('重置') }}</el-button>
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
  import flieUpload from '@/components/flieUpload/upLoad.vue';
  import setPassword from './setPassword.vue';
  import Terminal from '@/api/terminal.js';
  import IndexApi from '@/api/index.js';
  import autoTips from '@/components/autoTips/autoTips.vue';
  import { useUserStore } from '@/store/index';
  import SvgIcon from '@/components/svg-icon/SvgIcon.vue';
  import { DTime } from '@/utils/DateTime.js';
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const is_open_buffet = supplier.value?.is_open_buffet || 0;
  export default {
    components: { flieUpload, setPassword, autoTips, SvgIcon },
    data() {
      return {
        is_open_buffet: is_open_buffet,
        supplier: supplier,
        form: {
          carousel: [],
          is_auto_send: 0,
          auto_lock_screen: 300,
          language: [],
          default_language: null,
          advanced_password: false,
          server: {
            ip: '',
            port: '',
          },
          bind_list: [],

          is_buffet_order_limit: '1',
          buffet_order_limit: {
            is_limit_time: '1',
            limit_time: null,
            is_limit_num: '1',
            limit_num: null,
          },
          is_order_limit: '1',
          is_customer_order: '1',
          is_call_service: '1',
          is_voice_remind: '1',
          order_limit: {
            is_limit_time: '1',
            limit_time: null,
            is_limit_num: '1',
            limit_num: null,
          },
        },
        onlineList: [], //  在线设备列表
        offlineList: [], //离线设备列表
        origin: window.location.origin,
        port: window.location.port || '80',
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
      onSubmit() {
        let self = this;
        if (this.$refs['form-item']) {
          this.$refs['form-item'].validate();
        }
        setTimeout(() => {
          const errorItems = document.querySelectorAll('.el-form-item__error');
          if (errorItems.length > 0) {
            const firstErrorItem = errorItems[0];
            firstErrorItem.scrollIntoView({ behavior: 'smooth', block: 'center' });
          }
        }, 200);
        for (let i = 0; i < self.form.carousel.length; i++) {
          if (self.form.carousel[i].sort == null) {
            return;
          }
        }
        this.sortOne();
        let params = JSON.parse(JSON.stringify(self.form));
        //绑定的设备不用提清空
        params.bind_list = [];
        self.loading = true;
        Terminal.saveTablet(params, true)
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
        Terminal.getTablet()
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
      upLoad(data) {
        var type = '';
        if (data.file_type.includes('video')) {
          type = 'video';
        }
        if (data.file_type.includes('image')) {
          type = 'image';
        }
        this.form.carousel.push({
          real_name: data.real_name,
          file_path: data.file_path,
          sort: 0,
          type: type,
        });
      },
      deleteOne(scope) {
        if (this.form.carousel.length == 1) return;
        this.form.carousel.splice(scope.$index, 1);
        this.form.carousel.sort((a, b) => {
          return a.sort - b.sort; // 按照数值大小进行排序
        });
      },
      sortOne() {
        this.form.carousel.sort((a, b) => {
          return a.sort - b.sort; // 按照数值大小进行排序
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
