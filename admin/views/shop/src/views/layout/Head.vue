<template>
  <div class="common-header">
    <div class="breadcrumb">
      <!--一般的标题显示-->
      <div class="baseInfo-left-base d-s-c">
        <span class="name">
          <template v-if="cloudBasic?.base?.brand_logo_long">
            <img :src="cloudBasic?.base?.brand_logo_long" />
          </template>
          <template v-else>
            <img src="@/assets/TTPOS-logo.png" />
          </template>
        </span>
      </div>

      <div class="header-navbar">
        <div class="header-time" v-if="days <= reminder && day != 0">
          {{ $t('剩余时间') }}: <span>{{ days }}</span> {{ lastTime ? $t('小时') : $t('天') }}
        </div>
        <div class="header-lang">
          <el-dropdown trigger="click" @command="setLanguage">
            <span class="el-dropdown-link">
              <SvgIcon class="language-icon" name="language"></SvgIcon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <template v-for="item in languageList">
                  <el-dropdown-item v-if="item.name" :disabled="item.name == languageTag" :command="item.name">
                    <div class="language-div">{{ item.value }}<img v-if="item.name == languageTag" src="../../assets/img/Check.svg" /></div>
                  </el-dropdown-item>
                </template>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
        <!-- <div class="header-navbar-icon">
                    <span class="gray">当前版本：{{ userInfo.version }}</span>
                </div> -->
        <div class="header-navbar-icon">
          <span class="ml4 icon iconfont icon-geren9"></span>
          <span class="text ml4 blue">{{ userInfo.userName }}，{{ $t('欢迎您！') }}</span>
        </div>
        <div class="header-navbar-icon"><span class="gray">|</span></div>
        <div class="header-navbar-icon" @click="passwordFunc()">
          <span class="text">{{ $t('修改密码') }}</span>
        </div>
        <div class="header-navbar-icon login-out" @click="exit()">
          <span class="icon iconfont icon-tuichu"></span>
          <span class="text ml4">{{ $t('退出') }}</span>
        </div>
      </div>
    </div>

    <!--修改密码-->
    <UpdatePassword v-if="is_password" @close="closeFunc"></UpdatePassword>
  </div>
</template>

<script setup>
  import { reactive, toRefs, onBeforeUnmount } from 'vue';
  import { useRouter } from 'vue-router';
  import { useUserStore } from '@/store';
  import UpdatePassword from './part/UpdatePassword.vue';
  import { languageStore } from '@/store/model/language.js';
  import SvgIcon from '@/components/svg-icon/SvgIcon.vue';
  import { EEUIRELOAD } from '@/utils/platform.js';

  const router = useRouter();
  const { userInfo, bus_on, bus_emit, bus_off, afterLogout, computedSupplier } = useUserStore();
  const language = languageStore();
  const languageTag = languageStore().language;
  const languageList = language.getLanguageList().languageList;
  const cloudBasic = language.getCloudBasic().cloudBasic;

  const state = reactive({
    /*菜单名称*/
    menu_title: $t('菜单'),
    /*切换菜单*/
    tabList: [],
    /*切换选中*/
    activeValue: 0,
    /*是否修改密码*/
    is_password: false,
    /*tab切换类别*/
    tab_type: undefined,
  });

  const supplier = computedSupplier().supplier;
  const reminder = supplier.value.setting?.reminder || 0;
  const exp_time = supplier.value.exp_time - new Date().getTime() / 1000 || 0;
  const day = supplier.value.day || 0;
  const millisecondsInADay = 24 * 60 * 60;
  let days = Math.ceil(exp_time / millisecondsInADay);
  let lastTime = false;
  if (days == '1') {
    days = Math.ceil((exp_time / millisecondsInADay) * 24);
    lastTime = true;
  }

  // 事件总线订阅
  bus_on('MenuName', (res) => {
    state.menu_title = res;
  });
  bus_on('tabData', (res) => {
    state.tabList = res.list;
    state.activeValue = res.active;
    state.tab_type = res.tab_type;
  });
  bus_on('activeValue', (res) => {
    if (res && res.params) {
      state.activeValue = res.params;
    } else {
      state.activeValue = res;
    }
  });
  bus_on('noTarget', (res) => {
    state.activeValue = res;
  });

  // 发送头部加载完成事件
  bus_emit('headLoad', true);

  onBeforeUnmount(() => {
    // 反订阅，注意与订阅事件名称大小写一致
    bus_off('MenuName');
    bus_off('tabData');
    bus_off('activeValue');
    bus_off('noTarget');
  });

  // 退出登录
  const logout = async () => {
    await afterLogout();
    router.push('/login');
  };

  const exit = () => {
    ElMessageBox.confirm($t('此操作将退出登录, 是否继续?'), $t('提示'), {
      confirmButtonText: $t('确定'),
      cancelButtonText: $t('取消'),
      type: 'warning',
    })
      .then(async () => {
        await logout();
      })
      .catch(() => {
        ElMessage({
          type: 'info',
          message: $t('已取消退出'),
        });
      });
  };

  // 切换语言
  const setLanguage = (e) => {
    if (e == languageTag) return;
    ElMessageBox.confirm($t('切换语言需要刷新后生效，是否确定刷新?'), $t('提示'), {
      confirmButtonText: $t('确定'),
      cancelButtonText: $t('取消'),
      type: 'warning',
    })
      .then(() => {
        language.setLanguage(e);
        EEUIRELOAD();
      })
      .catch(() => {
        ElMessage({
          type: 'info',
          message: $t('已取消'),
        });
      });
  };

  // 修改密码弹窗
  const passwordFunc = () => {
    state.is_password = true;
  };
  const closeFunc = () => {
    state.is_password = false;
  };

  // 暴露到模板
  const { menu_title, tabList, activeValue, is_password, tab_type } = toRefs(state);
</script>

<style lang="scss">
  .common-header .el-tabs__nav-wrap::after {
    display: none;
  }

  .login-out .icon-tuichu {
    color: red;
  }

  .header-navbar-icon .icon-geren9 {
    font-size: 20px;
  }

  .language-icon {
    color: #fff;
    width: 16px;
    height: 16px;
  }

  .language-div {
    width: 90px;
    display: flex;
    align-items: center;
    justify-content: space-between;

    img {
      width: 18px;
    }
  }
</style>
