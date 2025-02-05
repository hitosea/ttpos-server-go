<template>
  <el-config-provider :locale="localeLanguage">
    <router-view />
  </el-config-provider>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import { getLanguage } from '@/i18n';
  import { useUserInfoStore } from '@/stores/userInfo';
  import dayjs from 'dayjs';
  // 日期组件翻译
  import 'dayjs/locale/zh-cn';
  import 'dayjs/locale/zh-tw';
  import 'dayjs/locale/en';
  import 'dayjs/locale/th'; // 泰语
  import 'dayjs/locale/ja'; // 日语
  import 'dayjs/locale/ko'; // 韩语
  // import 'dayjs/locale/tr'; // 土耳其语
  //
  import zhCN from 'element-plus/es/locale/lang/zh-cn';
  import zhTW from 'element-plus/es/locale/lang/zh-tw';
  import en from 'element-plus/es/locale/lang/en';
  import th from 'element-plus/es/locale/lang/th'; // 泰语
  import ja from 'element-plus/es/locale/lang/ja'; // 日语
  import ko from 'element-plus/es/locale/lang/ko'; // 韩语
  // import tr from 'element-plus/es/locale/lang/tr'; // 土耳其语
  import { v4 as uuidv4 } from 'uuid';

  const existingUuid = localStorage.getItem('uuid');
  if (!existingUuid) {
    const newUuid = uuidv4();
    localStorage.setItem('uuid', newUuid);
  }

  const localeLanguage = computed(() => {
    if (getLanguage() == 'zh') {
      dayjs.locale('zh-cn');
      return zhCN;
    }
    if (getLanguage() == 'zhtw') {
      dayjs.locale('zh-tw');
      return zhTW;
    }
    if (getLanguage() == 'th') {
      dayjs.locale('th');
      return th;
    }
    if (getLanguage() == 'ja') {
      dayjs.locale('ja');
      return ja;
    }
    if (getLanguage() == 'ko') {
      dayjs.locale('ko');
      return ko;
    }
    // if (getLanguage() == 'tr') {
    //   dayjs.locale('tr');
    //   return tr;
    // }
    dayjs.locale('en');
    return en;
  });
  //
  const { getSetting, getRoleList } = useUserInfoStore();
  // 权限信息
  getRoleList(true);
  // 基础信息
  getSetting(true);
</script>
