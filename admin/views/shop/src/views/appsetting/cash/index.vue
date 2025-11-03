<template>
  <div class="cash" v-loading="loading">
    <div class="form-div">
      <el-form size="small" ref="form" :model="form" label-position="top">
        <el-form-item for="no_click" :label="$t('轮播内容')" prop="carousel" :rules="[{ required: false, message: '' }]">
          <div class="draggable-list">
            <flieUpload
              @upLoad="upLoad"
              :source="'cashier'"
              :tips="[
                $t('图片：支持JPG、JPEG、PNG、WEBP格式，小于15MB'),
                $t('视频：支持MP4格式，小于30MB'),
                $t('不同设备尺寸：'),
                $t('【商米】：1024*600px'),
                '【COMPAX】：1280*800px',
              ]"
            >
            </flieUpload>
            <el-table size="small" :data="form.carousel" border v-loading="loading">
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
                    <el-input-number :controls="false" :min="0" :max="999" :placeholder="$t('接近0，排序等级越高')" v-model.number="scope.row.sort"></el-input-number>
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
                  <div class="delete-box">
                    <el-icon size="24">
                      <Delete @click="deleteOne(scope)" />
                    </el-icon>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('用餐方式')" prop="order_method" :rules="[{ required: true, message: '' }]">
          <el-checkbox
            v-model="form.order_method.is_cashier_order"
            :disabled="form.order_method.is_table_order == '0'"
            true-value="1"
            false-value="0"
            :label="$t('点餐')"
            size="large"
          ></el-checkbox>
          <el-checkbox
            v-model="form.order_method.is_table_order"
            :disabled="form.order_method.is_cashier_order == '0'"
            true-value="1"
            false-value="0"
            :label="$t('桌台')"
            size="large"
          ></el-checkbox>
        </el-form-item>

        <!-- <el-form-item for="no_click"  :label="$t('收银结账自动送厨房：')" prop="is_auto_send" :rules="[{ required: true, message: '' }]">
                <el-radio-group v-model="form.is_auto_send">
                    <el-radio label="1">{{ $t('开') }}</el-radio>
                    <el-radio label="0">{{ $t('关') }}</el-radio>
                </el-radio-group>
            </el-form-item> -->

        <el-form-item for="no_click" v-if="is_open_buffet" :label="$t('自助餐剩余时长颜色：')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_remain_color">
            <el-radio value="1">{{ $t('开') }}</el-radio>
            <el-radio value="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item for="no_click" v-if="form.is_remain_color == 1 && is_open_buffet" label="" :rules="[{ required: true, message: '' }]">
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
          <el-input class="max-w460" type="password" v-model="password" disabled></el-input>
          <el-button @click="setPassword($t('修改高级密码'))" type="primary" link size="small">{{ $t('设置密码') }}</el-button>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('钱箱密码')" prop="is_open_cashier_password" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_open_cashier_password">
            <el-radio value="1">{{ $t('开') }}</el-radio>
            <el-radio value="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item v-if="form.is_open_cashier_password == '1'" for="no_click" prop="password" :rules="[{ required: true, message: '', trigger: 'chenge' }]">
          <el-input class="max-w460" type="password" v-model="password" disabled></el-input>
          <el-button @click="setPassword($t('修改钱箱密码'))" type="primary" link size="small">{{ $t('设置密码') }}</el-button>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('自动锁屏')" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.is_auto_lock_screen">
            <el-radio value="1">{{ $t('开') }}</el-radio>
            <el-radio value="0">{{ $t('关') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <template v-if="form.is_auto_lock_screen == '1'">
          <el-form-item for="no_click" :label="$t('锁屏密码')" prop="password" :rules="[{ required: true, message: '', trigger: 'chenge' }]">
            <el-input class="max-w460" type="password" v-model="password" disabled></el-input>
            <el-button @click="setPassword($t('修改锁屏密码'))" type="primary" link size="small">{{ $t('设置密码') }}</el-button>
          </el-form-item>
          <el-form-item for="no_click" label="" prop="auto_lock_screen" :rules="[{ required: true, message: '' }]">
            <el-select v-model="form.auto_lock_screen" class="max-w460">
              <template v-for="(item, index) in lockList" :key="index">
                <el-option :value="item.key" :label="item.label">{{ item.label }}</el-option>
              </template>
            </el-select>
          </el-form-item>
        </template>

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
        <el-form-item for="no_click" :label="$t('已登录的设备') + `(${onlineList.length}/${form.licenses?.cash_limit})`" class="form-items">
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
            <div class="device-btn-box">
              <el-button @click="handleClick(item)" type="primary" link size="small">{{ $t('解绑') }}</el-button>
              <el-checkbox class="ml20" v-if="item.platform != 0" :model-value="item.is_main == 1" @change="setMain(item)" size="small">{{ $t('设为主收银机') }}</el-checkbox>
            </div>
          </div>
          <p v-else>{{ $t('暂无设备') }}</p>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('已退出登录的设备')" class="form-items">
          <div class="form-div" v-for="(item, index) in offlineList" v-if="offlineList.length > 0">
            <div class="max-w460 input-ss">
              <autoTips :tooltipMaxWidth="460" :content="(item.remark ? item.remark : '') + `(${item.key})`">{{ (item.remark ? item.remark : '') + `(${item.key})` }}</autoTips>
            </div>
            <div class="device-btn-box">
              <el-button @click="handleClick(item)" type="primary" link size="small">{{ $t('解绑') }}</el-button>
              <el-checkbox class="ml20" v-if="item.platform != 0" :model-value="item.is_main == 1" @change="setMain(item)" size="small">{{ $t('设为主收银机') }}</el-checkbox>
            </div>
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
      v-if="open"
      :open="open"
      :openTitle="openTitle"
      :cashierPassword="form.cashier_password"
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
        open: false,
        openTitle: '',
        loading: false,
        form: {
          server: {
            ip: '',
            port: 8080,
          },
          order_method: {
            is_cashier_order: '1',
            is_table_order: '1',
          },
          is_remain_color: '0',
          is_cashier_password: '1',
          is_open_cashier_password: '1',
          remain_color: ['', ''],
          carousel: [],
          is_auto_send: 1,
          is_auto_lock_screen: 1,
          auto_lock_screen: null,
          language: [],
          default_language: null,
          cashier_password: false,
          bind_list: [],
        },
        onlineList: [], //  在线设备列表
        offlineList: [], //离线设备列表
        origin: '',
        port: window.location.port || '80',
        input1: $t('10分钟以内'),
        input2: $t('20分钟以内'),
        password: '',
        languageList: [],
        lockList: [
          {
            label: $t('无操作15秒'),
            key: 15,
          },
          {
            label: $t('无操作30秒'),
            key: 30,
          },
          {
            label: $t('无操作1分钟'),
            key: 60,
          },
          {
            label: $t('无操作2分钟'),
            key: 120,
          },
          {
            label: $t('无操作5分钟'),
            key: 300,
          },
          {
            label: $t('无操作10分钟'),
            key: 600,
          },
        ],
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
        Terminal.getTerminal()
          .then((data) => {
            self.loading = false;
            self.form = data.data.vars.values;
            self.languageList = data.data.vars.values.language_list;
            if (self.form.cashier_password) {
              self.password = 666666;
            }
            //
            self.form.auto_lock_screen = Number(self.form.auto_lock_screen || 0);
            //
            self.form.language = self.form.language.filter((lang) => {
              return self.languageList.map((h) => h.key).indexOf(lang) != -1;
            });
            self.form.carousel = self.form.carousel.map((item) => {
              return {
                file_path: item.file_path,
                real_name: item.real_name,
                sort: Number(item.sort),
                type: item.type,
              };
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
        if (this.form.order_method.is_cashier_order == 0 && this.form.cashier_count > 0) {
          ElMessageBox.confirm($t('点餐正在使用，是否确认取消对应订单，并关闭用餐方式'), $t('提示'), {
            confirmButtonText: $t('确定'),
            cancelButtonText: $t('取消'),
            type: 'warning',
          })
            .then(() => {
              this.save();
            })
            .catch(() => {});
        } else if (this.form.order_method.is_table_order == 0 && this.form.table_count > 0) {
          ElMessageBox.confirm($t('桌台正在使用，是否确认取消对应订单，并关闭用餐方式'), $t('提示'), {
            confirmButtonText: $t('确定'),
            cancelButtonText: $t('取消'),
            type: 'warning',
          })
            .then(() => {
              this.save();
            })
            .catch(() => {});
        } else {
          this.save();
        }
      },

      save() {
        let self = this;
        if (this.$refs['form-item']) {
          this.$refs['form-item'].validate();
        }
        setTimeout(() => {
          const errorItems = document.querySelectorAll('.el-form-item__error');
          if (errorItems.length > 0) {
            const firstErrorItem = errorItems[0];
            firstErrorItem.scrollIntoView({
              behavior: 'smooth',
              block: 'center',
            });
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
        params.main_cashier_uuid = params.bind_list.find((bind) => bind.is_main == 1)?.uuid || '';
        params.bind_list = [];
        self.loading = true;
        Terminal.saveTerminal(params, true)
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
      },

      //解绑
      handleClick(item) {
        let self = this;
        let message = $t('解绑后不可恢复，确认解绑吗?');
        if (item.is_cashier_shift == 0) {
          message = $t('当前收银机尚未交班，确认解绑吗?');
        }

        ElMessageBox.confirm(message, $t('提示'), {
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
      setMain(item) {
        let self = this;
        self.form.bind_list.map((bind) => {
          if (bind.id != item.id) {
            bind.is_main = 0;
          }
        });
        item.is_main = item.is_main == 1 ? 0 : 1;
      },
    },
  };
</script>
<style scoped lang="scss">
  .cash {
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

  .color-box {
    display: flex;
    gap: 12px;
    margin-right: 16px;
  }

  .draggable-list {
    width: 100%;
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
    .device-btn-box {
      width: 660px;
    }
  }
</style>
