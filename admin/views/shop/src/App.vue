<template>
  <el-config-provider :locale="locale">
    <router-view />
    <div class="expire-body" v-if="expireShow">
      <div class="expire-box">
        <img src="./assets/img/expire.svg" />
        <h3 class="expire-h3">{{ $t('店铺到期') }}</h3>
        <p class="expire-p">
          {{ $t('店铺状态已到期，如需继续使用，请联系销售代表') }}
        </p>
        <!-- <el-form size="small" style="width: 100%" ref="rulesForm" :model="form" label-position="top">
          <el-form-item prop="authorization_code" :rules="[{ required: true, message: $t('请输入授权码') }]">
            <el-input type="text" :loading="loading" :placeholder="$t('请输入授权码')" v-model="form.authorization_code" class="max-w460"></el-input>
          </el-form-item>
        </el-form> -->
        <el-button class="expire-ok" size="small" type="primary" @click="handleRefresh()" :loading="loading">{{ $t('刷新') }}</el-button>
      </div>
    </div>
    <el-dialog v-model="accreditShow" :title="$t('设备授权')" width="720" :before-close="handleClose" align-center>
      <div class="accredit-main">
        <p>{{ $t('请根据以下支持的设备数量上限调整绑定的设备') }}</p>
        <div class="accredit-num">
          <p>
            {{ $t('收银机支持') }}<span>{{ verifyAuthData?.c_l }}</span>
          </p>
          <p>
            {{ $t('厨显支持') }}<span>{{ verifyAuthData?.k_l }}</span>
          </p>
          <p>
            {{ $t('平板支持') }}<span>{{ verifyAuthData?.t_l }}</span>
          </p>
        </div>
        <p class="accredit-title">{{ $t('以下为收银机当前绑定设备') }}</p>
        <template v-for="item in terminal">
          <div class="accredit-one" v-if="item.source == 'cashier'">
            <div class="input-ss">
              <autoTips :content="(item.remark ? item.remark : '') + `(${item.key})`">{{ (item.remark ? item.remark : '') + `(${item.key})` }}</autoTips>
            </div>
            <el-button :loading="loading" @click="handleClick(item.id)">{{ $t('解绑') }}</el-button>
          </div>
        </template>

        <p class="accredit-title">{{ $t('以下为厨显当前绑定设备') }}</p>
        <template v-for="item in terminal">
          <div class="accredit-one" v-if="item.source == 'kitchen'">
            <div class="input-ss">
              <autoTips :content="(item.remark ? item.remark : '') + `(${item.key})`">{{ (item.remark ? item.remark : '') + `(${item.key})` }}</autoTips>
            </div>
            <el-button :loading="loading" @click="handleClick(item.id)">{{ $t('解绑') }}</el-button>
          </div>
        </template>

        <p class="accredit-title">{{ $t('以下为平板当前绑定设备') }}</p>
        <template v-for="item in terminal">
          <div class="accredit-one" v-if="item.source == 'tablet'">
            <div class="input-ss">
              <autoTips :content="(item.remark ? item.remark : '') + `(${item.key})`">{{ (item.remark ? item.remark : '') + `(${item.key})` }}</autoTips>
            </div>
            <el-button :loading="loading" @click="handleClick(item.id)">{{ $t('解绑') }}</el-button>
          </div>
        </template>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button :loading="loading" @click="handleClose">{{ $t('取消') }}</el-button>
          <el-button :loading="loading" type="primary" @click="handleSubmit(2)"> {{ $t('确定') }}</el-button>
        </div>
      </template>
    </el-dialog>
  </el-config-provider>
</template>

<script setup>
  import { ref, reactive, computed, onMounted, onUnmounted } from 'vue';
  import { ElConfigProvider, ElMessageBox } from 'element-plus';
  import { useRoute, useRouter } from 'vue-router';
  import IndexApi from '@/api/index.js';
  import { useLockscreenStore } from './store/model/lockscreen';
  import { languageStore } from './store/model/language';
  import { useUserStore } from '@/store';
  import zhCn from 'element-plus/es/locale/lang/zh-cn';
  import zhTw from 'element-plus/es/locale/lang/zh-tw';
  import en from 'element-plus/es/locale/lang/en';
  import th from 'element-plus/es/locale/lang/th';
  import ko from 'element-plus/es/locale/lang/ko';
  import ja from 'element-plus/es/locale/lang/ja';
  import tr from 'element-plus/es/locale/lang/tr';
  import sv from 'element-plus/es/locale/lang/sv';
  import my from '@/lang/component/my';
  import { message } from '@/utils/message.js';
  import autoTips from '@/components/autoTips/autoTips.vue';
  import { getSessionStorage, setSessionStorage } from '@/utils/base.js';
  import configObj from '@/config';
  import { getStorage, setStorage } from '@/utils/storageData';
  import { createdAuth } from '@/utils/createdAuth.js';
  import { EEUIRELOAD } from '@/utils/platform.js';
  import { v4 as uuidv4 } from 'uuid';
  import { handRouterTable, handMenuData } from '@/utils/router';
  import AuthApi from '@/api/auth.js';

  const existingUuid = localStorage.getItem('uuid');
  if (!existingUuid) {
    const newUuid = uuidv4();
    localStorage.setItem('uuid', newUuid);
  }

  const { menu } = configObj;
  const { userInfo, changeUserInfo, setMenus, setRenderMenus, menus } = useUserStore();

  const useLockscreen = useLockscreenStore();
  const expireShow = computed(() => useLockscreen.expire);
  // const isLock = computed(() => useLockscreen.isLock);
  // const lockTime = computed(() => useLockscreen.lockTime);

  const language = ref(languageStore().getLanguageKey().language);

  const rulesForm = ref(null);
  const form = ref({
    authorization_code: '',
  });
  const verifyAuthData = ref('');
  const terminal = ref('');
  const accreditShow = ref(false);
  const loading = ref(false);
  const route = useRoute();

  const locale = ref(zhCn);
  if (language.value == 'zh') {
    locale.value = zhCn;
  }
  if (language.value == 'zhtw') {
    locale.value = zhTw;
  }
  if (language.value == 'en') {
    locale.value = en;
  }
  if (language.value == 'th') {
    locale.value = th;
  }
  if (language.value == 'ko') {
    locale.value = ko;
  }
  if (language.value == 'ja') {
    locale.value = ja;
  }
  if (language.value == 'my') {
    locale.value = my;
  }
  if (language.value == 'tr') {
    locale.value = tr;
  }
  if (language.value == 'sv') {
    locale.value = sv;
  }
  const state = reactive({});
  // let timer;
  // const timekeeping = () => {
  //     clearInterval(timer);
  //     if (route.name == 'login' || isLock.value) return;
  //     // 设置不锁屏
  //     useLockscreen.setLock(false);
  //     // 重置锁屏时间
  //     useLockscreen.setLockTime();
  //     timer = setInterval(() => {
  //         // 锁屏倒计时递减
  //         useLockscreen.setLockTime(lockTime.value - 1);
  //         if (lockTime.value <= 0) {
  //             // 设置锁屏
  //             useLockscreen.setLock(true);
  //             router.push('/lockscreen')
  //             return clearInterval(timer);
  //         }
  //     }, 1000);
  // };

  //刷新
  const handleRefresh = () => {
    if (route.path == '/login') {
      useLockscreen.setExpire(false);
    } else {
      useLockscreen.setExpire(false);
      EEUIRELOAD();
    }
  };

  //授权
  const handleSubmit = (e) => {
    if (e == 1) {
      rulesForm.value.validate((valid) => {
        if (valid) {
        }
      });
      if (!form.value.authorization_code) {
        return;
      }
    }
    loading.value = true;
    let params = {
      auth_code: form.value.authorization_code,
    };
    IndexApi.verifyAuthCode(params, true)
      .then((res) => {
        if (res.code == -102) {
          message({
            message: $t('授权码不正确，请联系销售代表'),
            type: 'error',
          });
        }
        if (res.data.upper_limit == 1 && res.code == 1) {
          verifyAuthData.value = res.data;
          if (e == 1) {
          } else {
            message({
              message: $t('请根据以下支持的设备数量上限调整绑定的设备'),
              type: 'warning',
            });
          }
          getBindData(e);
        }
        if (res.data.upper_limit == 0 && res.code == 1) {
          form.value.authorization_code = '';
          message({
            message: $t('操作成功'),
            type: 'success',
          });
          getBaes('refresh');
        }
        loading.value = false;
      })
      .catch((error) => {
        loading.value = false;
      });
  };

  const handleClose = () => {
    EEUIRELOAD();
  };

  /*获取绑定设备列表*/
  const getBindData = (e) => {
    loading.value = true;
    let params = {
      source: 'all',
    };
    IndexApi.getBindList(params, true)
      .then((res) => {
        terminal.value = res.data;
        if (e == 1) {
          accreditShow.value = true;
        }
        loading.value = false;
      })
      .catch((error) => {
        loading.value = false;
      });
  };

  //解绑
  const handleClick = (id) => {
    ElMessageBox.confirm($t('解绑后不可恢复，确认解绑吗?'), $t('提示'), {
      confirmButtonText: $t('确定'),
      cancelButtonText: $t('取消'),
      type: 'warning',
    })
      .then(() => {
        loading.value = true;
        let params = {
          bind_id: id,
        };
        IndexApi.unbind(params, true)
          .then((data) => {
            loading.value = false;
            message({
              message: $t('保存成功'),
              type: 'success',
            });
            getBindData();
          })
          .catch((error) => {
            loading.value = false;
          });
      })
      .catch(() => {});
  };

  // 获取基础信息
  const getBaes = (event) => {
    IndexApi.base(true)
      .then((res) => {
        if (res.code == -102) return;
        languageStore().setLanguageList(res.data.language);
        //获取是否在云端
        languageStore().setIsCloudDeploy(res.data.isCloudDeploy);
        //设置logo
        languageStore().setCloudBasic(res.data.cloudBasic);
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
        //判断后台是否删除了这个语言
        let langKeyArr = [];
        languageStore()
          .getLanguageList()
          .languageList.value.map((item) => {
            langKeyArr.push(item.name);
          });
        if (!langKeyArr.includes(languageStore().getLanguageKey().language.value)) {
          languageStore()
            .getLanguageList()
            .languageList.value.some((item) => {
              if (item.name) {
                languageStore().setLanguage(item.name);
                EEUIRELOAD();
                return true;
              }
            });
        }
        //设置授权数据
        setSessionStorage('supplier', res.data.supplier);
        // 设置erp数据
        setSessionStorage('erp', res.data.erp);
        /*获取基础配置*/
        const dataInfo = {
          data: {
            shop_name: res.data.settings.shop_name,
            logoUrl: res.data.settings.shop_bg_img,
            is_open_tax: res.data.settings.is_open_tax,
          },
        };
        changeUserInfo(dataInfo);
        let auth = getSessionStorage('authlist');
        let authlist = {};
        auth = getStorage(menu);
        createdAuth(auth, authlist);
        setSessionStorage('authlist', authlist);
        auth = authlist;
        if (event == 'refresh') {
          setTimeout(() => {
            EEUIRELOAD();
          }, 1000);
        }
      })
      .catch((error) => {});
  };

  const refreshRouter = async () => {
    try {
      const result = await AuthApi.getRoleList({ token: userInfo.token });
      let renderMenusList = handMenuData(JSON.parse(JSON.stringify(result.data.menus)));
      let menusList = handRouterTable(JSON.parse(JSON.stringify(result.data.menus)));
      let appId = userInfo.AppID;
      renderMenusList.forEach((item) => {
        item.path = appId + item.path;
        item.redirect_name && (item.redirect_name = '/' + appId + item.redirect_name);
        item.children?.forEach((child) => {
          child.path = '/' + appId + child.path;
          child.redirect_name && (child.redirect_name = '/' + appId + child.redirect_name);
          child.children?.forEach((childItem) => {
            childItem.path = '/' + appId + childItem.path;
            childItem.redirect_name && (childItem.redirect_name = '/' + appId + childItem.redirect_name);
          });
        });
      });
      //
      setMenus(menusList);
      setRenderMenus(renderMenusList);
      if (menus.length && menus.length != menusList.length) {
        EEUIRELOAD();
      }
    } catch (error) {
      console.log(error);
    }
  };

  onMounted(async () => {
    if (userInfo) {
      await refreshRouter();
      await getBaes();
    }
  });

  onUnmounted(() => {});
</script>

<style lang="scss">
  @import '@/assets/font/iconfont.css';
  @import '@/assets/font/myIcon.css';
  @import '@/styles/diy.css';

  * {
    margin: 0;
    padding: 0;
  }

  .common-level-rail {
    text-align: right;

    &.flex {
      display: flex;
      justify-content: space-between;
      margin-bottom: 0;
    }
  }

  .common-search-wrap {
    &.flex {
      display: flex;
      justify-content: space-between;
      margin-bottom: 0;
    }
  }

  .expire-body {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(16, 10, 5, 0.3);
    /* 半透明黑色遮罩层 */
    z-index: 999;
    /* 设置z-index使遮罩层位于其他内容上方 */
    display: flex;
    justify-content: center;
    align-items: center;

    .expire-box {
      padding: 32px 20px 20px 20px;
      background: #fff;
      border-radius: 4px;
      display: flex;
      width: 400px;
      flex-direction: column;
      align-items: center;

      .expire-h3 {
        color: var(--el-color-black);
        text-align: center;
        font-size: 18px;
        font-style: normal;
        font-weight: 700;
        line-height: normal;
        margin-top: 16px;
      }

      .expire-p {
        color: var(--el-color-black);
        text-align: center;
        font-size: 14px;
        font-style: normal;
        font-weight: 400;
        line-height: normal;
        text-transform: capitalize;
        margin-top: 8px;
        margin-bottom: 16px;
      }

      .expire-ok {
        margin-top: 16px;
        margin-left: auto;
      }
    }
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

      .time {
        flex-shrink: 0;
      }
    }
  }
</style>
