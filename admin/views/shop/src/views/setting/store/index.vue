<template>
  <!--
      时间：2019-10-26
      描述：设置-商城设置
  -->
  <div class="product-add" v-loading="loadingAuth">
    <!--form表单-->
    <el-form class="product-form" size="small" ref="form" :model="form" label-position="top" label-width="150px">
      <!--添加门店-->
      <el-form-item for="no_click" class="max-w460" :label="$t('店铺名称')" :rules="[{ required: true, message: ' ' }]" prop="name">
        <el-input v-model="form.name" :placeholder="$t('商城名称')" :maxlength="100" class="max-w460"></el-input>
      </el-form-item>
      <el-form-item for="no_click" class="max-w460" label="LOGO" :rules="[{ required: true, message: ' ' }]" prop="name">
        <div class="ww100">
          <el-button @click="chooseImg('logoUrl')">{{ $t('选择图片') }}</el-button>
        </div>
        <img class="mt10" v-img-url="form.logoUrl" :width="100" />
        <div class="tips">
          {{ $t('支持JPG、JPEG、PNG、WEBP格式，小于15MB，尺寸：120*120px') }}
        </div>
      </el-form-item>

      <!-- <el-form-item for="no_click" class="max-w460" :label="$t('抹零方式')" prop="address" :rules="[{ required: true, message: $t('请选择抹零方式') }]">
                <el-select v-model="form.zeroing_method" clearable size="default">
                    <el-option value="0" :label="$t('不抹零')">{{ $t('不抹零') }}</el-option>
                    <el-option value="1" :label="$t('抹分')">{{ $t('抹分') }}</el-option>
                    <el-option value="2" :label="$t('抹角')">{{ $t('抹角') }}</el-option>
                    <el-option value="3" :label="$t('四舍五入到角')">{{ $t('四舍五入到角') }}</el-option>
                    <el-option value="4" :label="$t('四舍五入到元')">{{ $t('四舍五入到元') }}</el-option>
                </el-select>
            </el-form-item> -->
      <el-form-item for="no_click" class="max-w460" :label="$t('白名单IP')">
        <el-input show-word-limit v-model="form.ip_white_list" :placeholder="$t('无')" :maxlength="500" disabled class="max-w460"></el-input>
      </el-form-item>

      <el-form-item for="no_click" class="max-w460" :label="$t('时区')" prop="time_zone" :rules="[{ required: true, message: $t('请选择时区') }]">
        <el-select v-model="form.time_zone" clearable size="default">
          <el-option v-for="item in form.time_zone_list" :value="item.key" :label="item.name">{{ item.name }}</el-option>
        </el-select>
      </el-form-item>

      <!-- <el-form-item for="no_click" :label="$t('结账后不清台')" prop="no_clear_table" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.no_clear_table">
          <el-radio label="0">{{ $t('清台') }}</el-radio>
          <el-radio label="1">{{ $t('不清台') }}</el-radio>
        </el-radio-group>
        <div class="tips">
          {{ $t('注：开启后结账不自动清台，结账后平板/H5再无法下单（对收银机/点餐助手）') }}
        </div>
      </el-form-item> -->
      <el-form-item for="no_click" class="max-w460" :label="$t('公司名称')">
        <el-input show-word-limit v-model="form.company" :placeholder="$t('请输入')" :maxlength="500" class="max-w460"></el-input>
      </el-form-item>

      <el-form-item for="no_click" class="max-w460" :label="$t('地址')">
        <el-input type="textarea" rows="3" show-word-limit v-model="form.address" :placeholder="$t('请输入')" :maxlength="500" class="max-w460"></el-input>
      </el-form-item>
      <el-form-item for="no_click" class="max-w460" :label="$t('经纬度')">
        <el-input v-model="form.coordinates" type="textarea" rows="3" :placeholder="$t('如：13.716412789763694, 100.52312952599786')" class="max-w460"></el-input>
      </el-form-item>

      <el-form-item for="no_click" class="max-w460" :label="$t('联系电话')" prop="phone" :rules="[{ required: true, message: $t('请输入联系电话') }]">
        <el-input v-model="form.phone" :placeholder="$t('请输入')" :maxlength="20" class="max-w460"></el-input>
      </el-form-item>

      <el-form-item for="no_click" class="max-w460" :label="$t('税号')">
        <el-input v-model="form.tax_number" :placeholder="$t('请输入')" :maxlength="255" class="max-w460"></el-input>
      </el-form-item>
      <el-form-item for="no_click" class="max-w460" :label="$t('店铺ID')" prop="customer">
        <el-input v-model="form.shop.shop_supplier_id" disabled placeholder="" class="max-w460"></el-input>
      </el-form-item>

      <template v-for="(item, index) in form.language" :key="index">
        <el-form-item
          for="no_click"
          class="lang-box"
          :label="$t('语言') + (index + 1) + (index == 0 ? '(' + $t('默认') + ')' : '')"
          :prop="`language[${index}].name`"
          :rules="[
            {
              required: true,
              validator: () => {
                return item.name ? true : false;
              },
              message: $t('请选择语言'),
            },
          ]"
        >
          <el-select v-model="item.name" clearable class="max-w460" @change="selectChange" :placeholder="$t('请选择语言')" size="default">
            <template v-for="cat in langList" :key="cat.name">
              <el-option :value="cat.name" :label="cat.value" :disabled="selectOne(cat.name)"></el-option>
            </template>
          </el-select>
          <el-icon size="24" class="delete-icon" :class="form.language.length < 2 ? 'no-click' : ''">
            <Delete @click="deleteOne(index)" />
          </el-icon>
        </el-form-item>
      </template>
      <el-form-item for="no_click">
        <el-button type="primary" @click="addOne" :loading="loading">{{ $t('添加语言') }}</el-button>
      </el-form-item>

      <el-form-item for="no_click" class="max-w460" :label="$t('客户端版本')" prop="customer">
        {{ `${client_version}` }}
        <el-button v-if="IsEeui" style="margin-left: 8px" :loading="updateLoading" @click="checkNew" type="primary">{{ $t('检查新版本') }}</el-button>
      </el-form-item>
      <el-form-item for="no_click" class="max-w460" :label="$t('服务端版本')" prop="customer">
        {{ version.server_version }}
      </el-form-item>
    </el-form>
    <!--提交-->
    <div class="common-button-wrapper">
      <el-button @click="getParams" :loading="loading">{{ $t('重置') }}</el-button>
      <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('保存') }}</el-button>
    </div>
    <!--上传图片-->
    <Upload v-if="isupload" :isupload="isupload" :type="type" :config="{ total: 1 }" @returnImgs="returnImgsFunc"> </Upload>

    <el-dialog v-model="newShow" :title="$t('新版本')" width="420" align-center :close-on-click-modal="false">
      <div class="mb-8 updata-box">
        <div class="updata-box-l">{{ $t('版本号') }}:</div>
        <div class="updata-box-r">{{ version_name }}</div>
      </div>
      <div class="updata-box">
        <div class="updata-box-l">{{ $t('更新内容') }}:</div>
        <div class="updata-box-r">{{ update_log }}</div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button
            :loading="loading"
            v-if="forced_update == '0'"
            @click="
              () => {
                newShow = false;
              }
            "
            >{{ $t('暂不更新') }}</el-button
          >
          <el-button type="primary" :loading="loading" @click="handleUpdata"> {{ $t('更新') }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script>
  import SettingApi from '@/api/setting.js';
  import IndexApi from '@/api/index.js';

  import Upload from '@/components/file/Upload.vue';
  import { languageStore } from '@/store/model/language';
  import { useUserStore } from '@/store';
  import autoTips from '@/components/autoTips/autoTips.vue';
  import { getSessionStorage, setSessionStorage } from '@/utils/base.js';
  import { createdAuth } from '@/utils/createdAuth.js';
  import { getStorage } from '@/utils/storageData';
  import { EEUIRELOAD, ISEEUI } from '@/utils/platform.js';
  import configObj from '@/config';
  //
  const { computedSupplier, computedUserInfo, changeUserInfo } = useUserStore();
  const userInfo = computedUserInfo().userInfo;
  const IsEeui = ISEEUI();
  const supplier = computedSupplier().supplier;
  const exp_time = supplier.value.exp_time - new Date().getTime() / 1000 || 0;
  const day = supplier.value.day || 0;
  const { menu } = configObj;
  let millisecondsInADay = 24 * 60 * 60;
  let days = Math.ceil(exp_time / millisecondsInADay);
  let lastTime = false;
  if (days == '1') {
    days = Math.ceil((exp_time / millisecondsInADay) * 24);
    lastTime = true;
  }

  export default {
    components: {
      Upload,
      autoTips,
    },
    data() {
      return {
        IsEeui,
        newShow: false,
        download_url: '',
        update_log: '',
        version_name: '',
        forced_update: '0',
        /*是否正在加载*/
        loading: false,
        loadingAuth: false,
        /*form表单数据*/
        form: {
          name: '',
          customer: '',
          address: '',
          coordinates: '',
          phone: '',
          tax_number: '',
          key: '',
          is_get_log: 0,
          avatarUrl: '',
          shop: {},
          zeroing_method: '0',
          company: '',
          time_zone: '',
          no_clear_table: '0',
          ip_white_list: '',
          language: [],
        },
        version: {},
        client_version: import.meta.env.VITE_BASIC_VERSION,
        code: '',
        all_type: [],
        type: [],
        /*是否打开图片选择*/
        isupload: false,
        langList: [],
        days: days,
        day: day,
        userInfo: userInfo,
        lastTime: lastTime,
        terminal: [],
        verifyAuthData: '',
        menu: menu,
      };
    },
    created() {
      this.getParams();
    },

    methods: {
      selectOne(lang) {
        let result = false;
        this.form.language.map((item) => {
          if (lang == item.name) {
            result = true;
          }
        });
        return result;
      },

      selectChange() {
        this.form.language.map((item) => {
          this.langList.map((items) => {
            if (item.name == items.name) {
              item.value = items.value;
            }
          });
        });
      },

      /*获取配置数据*/
      getParams() {
        let self = this;
        self.loading = true;
        SettingApi.storeDetail({}, true)
          .then((res) => {
            self.loading = false;
            let vars = res.data.vars.values;
            // self.form = formatModel(self.form, vars);
            self.form = Object.assign(self.form, vars);
            self.form.language = vars.language || [];
            self.version = res.data.version;
            self.form.shop = res.data.shop;

            self.langList = res.data.shop.language || [];
            let langArr = [];
            self.langList.map((item) => {
              langArr.push(item.name);
            });
            self.form.language.map((item) => {
              if (langArr.indexOf(item.name) == -1) {
                item.name = '';
                item.value = '';
              }
            });
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      /*提交*/
      onSubmit() {
        let self = this;
        let params = this.form;
        params.language.map((item, index) => {
          item.key = index + 1;
        });
        self.$refs.form.validate((valid) => {
          if (valid) {
            self.loading = true;
            SettingApi.editStore(params, true)
              .then((data) => {
                self.loading = false;
                let nowLanguage = JSON.parse(localStorage.getItem('Language')).language;
                let lang = [];
                self.form.language.map((item) => {
                  lang.push(item.name);
                });
                if (lang.indexOf(nowLanguage) == -1) {
                  languageStore().setLanguage(self.form.language[0].name);
                }

                languageStore().setLanguageList(params.language);
                this.$ElMessage({
                  message: $t('操作成功'),
                  type: 'success',
                });
                setTimeout(() => {
                  EEUIRELOAD();
                }, 1000);
              })
              .catch((error) => {
                self.loading = false;
              });
          }
        });
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
      },
      /*选择图片*/
      chooseImg(e) {
        this.type = e;
        this.isupload = true;
      },
      /*关闭选择图片*/
      returnImgsFunc(e) {
        this.isupload = false;
        if (e != null && e.length > 0) {
          if (this.type == 'avatarUrl') {
            this.form.avatarUrl = e[0].file_path;
          } else if (this.type == 'logoUrl') {
            this.form.logoUrl = e[0].file_path;
          }
        }
      },

      //重新绑定先验证授权码
      handleAuth(e) {
        if (!this.code && e == 1) {
          this.$ElMessage({
            message: $t('请输入授权码'),
            type: 'warning',
          });
          return;
        }

        let self = this;
        self.loading = true;
        let params = {
          auth_code: self.code,
        };
        IndexApi.verifyAuthCode(params, true)
          .then((res) => {
            if (res.data.upper_limit == 1) {
              self.verifyAuthData = res.data;
              if (e == 1) {
                self.dialogVisible = false;
              } else {
                this.$ElMessage({
                  message: $t('请根据以下支持的设备数量上限调整绑定的设备'),
                  type: 'warning',
                });
              }
              self.getBindData(e);
            } else {
              self.code = '';
              self.dialogVisible = false;
              this.$ElMessage({
                message: $t('操作成功'),
                type: 'success',
              });
              self.loadingAuth = true;
              self.getBase();
            }
            self.loading = false;
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      //检查新版本
      checkNew() {
        let self = this;
        self.updateLoading = true;
        let brand = 2;
        if (window.config instanceof Function && window.config().brand == 'TTPOS') {
          brand = 1;
        }

        SettingApi.getNewVersion({ brand }, true)
          .then((data) => {
            self.updateLoading = false;
            if (self.isAndroidInside()) {
              let eeui = requireModuleJs('eeui');
              let version = eeui.getLocalVersion();
              if (Number(data.data.version_number) > version) {
                self.download_url = data.data.download_url;
                self.update_log = data.data.update_log;
                self.version_name = data.data.version_name;
                self.forced_update = data.data.forced_update;
                self.newShow = true;
              } else {
                this.$ElMessage({
                  message: $t('已经是最新版本'),
                  type: 'success',
                });
              }
            }
          })
          .catch((error) => {
            self.updateLoading = false;
          });
      },

      //确定更新
      handleUpdata() {
        let eeui = requireModuleJs('eeui');
        eeui.openWeb(this.download_url);
      },

      //判断是不是APP
      isAndroidInside() {
        if (/android_kuaifan_eeui/.test(navigator.userAgent)) {
          return true;
        } else {
          return false;
        }
      },

      // 获取基础信息
      getBase() {
        IndexApi.base(true)
          .then(async (res) => {
            languageStore().setLanguageList(res.data.language);
            const data = {};
            res.data.language.map((item) => {
              data[item.key] = '';
            });
            languageStore().setLanguageData(data);
            //刷新
            let language = JSON.parse(localStorage.getItem('Language'));
            if (!language) {
              EEUIRELOAD();
            }
            //判断默认语言
            if (language && language.language == '' && language.languageList[0]?.name) {
              languageStore().setLanguage(language.languageList[0]?.name);
            }
            /*获取基础配置*/
            const dataInfo = {
              data: {
                shop_name: res.data.settings.shop_name,
                logoUrl: res.data.settings.shop_bg_img,
                is_open_tax: res.data.settings.is_open_tax,
              },
            };
            //设置授权数据
            setSessionStorage('supplier', res.data.supplier);
            // 设置erp数据
            setSessionStorage('erp', res.data.erp);
            // 设置settings数据
            setSessionStorage('settings', res.data.settings);
            await changeUserInfo(dataInfo);
            let auth = getSessionStorage('authlist');
            let authlist = {};
            auth = getStorage(this.menu);
            createdAuth(auth, authlist);
            setSessionStorage('authlist', authlist);
            auth = authlist;

            //获取完再跳转
            setTimeout(() => {
              EEUIRELOAD();
            }, 1000);
          })
          .catch((error) => {});
      },

      handleCloseDia() {
        this.dialogVisible = false;
        this.code = '';
      },

      deleteOne(index) {
        if (this.form.language.length < 2) return;
        this.form.language.splice(index, 1);
      },

      addOne() {
        this.form.language.push({
          key: this.form.language.length + 1,
          name: '',
          value: '',
        });
      },
    },
  };
</script>
<style scoped lang="scss">
  .product-add {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    .product-form {
      flex: 1 1 auto;
      overflow-y: auto;
    }

    .common-button-wrapper {
      flex: 0 0 auto;
      flex-shrink: 1;
    }
  }
  .tips {
    color: #ccc;
  }

  input::-webkit-outer-spin-button,
  input::-webkit-inner-spin-button {
    -webkit-appearance: none;
  }

  input[type='number'] {
    -moz-appearance: textfield;
  }

  .time {
    margin: 0 16px;
  }

  .accredit-main {
  }

  .accredit-num {
    display: flex;
    margin-top: 8px;
    gap: 16px;

    span {
      color: #ff0000;
      margin-left: 4px;
      font-weight: bold;
    }
  }

  .accredit-title {
    margin-top: 16px;
  }

  .accredit-one {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 8px;
    overflow: hidden;
  }

  .input-ss {
    background-color: var(--el-disabled-bg-color);
    box-shadow: 0 0 0 1px var(--el-disabled-border-color) inset;
    background-image: none;
    border-radius: var(--el-input-border-radius, var(--el-border-radius-base));
    cursor: text;
    transition: var(--el-transition-box-shadow);
    transform: translate3d(0, 0, 0);
    width: 100%;
    overflow: hidden;
    height: 32px;
    padding: 6px 8px;
    position: relative;
  }

  .no-wrap {
    :deep(.el-form-item__content) {
      flex-wrap: nowrap;
      gap: 8px;

      .time {
        flex-shrink: 0;
      }
    }
  }

  .flex {
    display: flex;
    width: 100%;
    gap: 8px;
    align-items: center;

    img {
      width: 20px;
      height: 20px;
      cursor: pointer;
    }
  }

  .lang-box {
    :deep(.el-form-item__content) {
      display: flex;
      width: 100%;
      gap: 12px;

      .delete-icon {
        cursor: pointer;
      }

      .no-click {
        color: #999;
        cursor: not-allowed;
      }
    }
  }
  .mb-8 {
    margin-bottom: 8px;
  }
  .updata-box {
    display: flex;
    .updata-box-l {
      width: 120px;
      flex-shrink: 0;
    }
    .updata-box-r {
      overflow-wrap: anywhere;
      white-space: pre-line;
    }
  }
</style>
