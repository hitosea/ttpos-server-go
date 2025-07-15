import { defineStore } from 'pinia';
import { computed } from 'vue';
export const languageStore = defineStore({
  id: 'Language',
  state: () => ({
    language: 'en',
    languageList: [
      {
        key: 'en',
        name: 'en',
        value: 'English',
      },
      {
        key: 'zh',
        name: 'zh',
        value: '简体中文',
      },
      {
        key: 'zhtw',
        name: 'zhtw',
        value: '繁體中文',
      },
      {
        key: 'th',
        name: 'th',
        value: 'ภาษาไทย',
      },
      {
        key: 'ja',
        name: 'ja',
        value: '日本語です',
      },
      {
        key: 'ko',
        name: 'ko',
        value: '한국어',
      },
      {
        key: 'tr',
        name: 'tr',
        value: 'Türkçe',
      },
      {
        key: 'my',
        name: 'my',
        value: 'မြန်မာဘာသာ',
      },
      {
        key: 'sv',
        name: 'sv',
        value: 'Svenska',
      },
    ],
    languageListOrigin: [
      {
        key: 'en',
        name: 'en',
        value: 'English',
      },
      {
        key: 'zh',
        name: 'zh',
        value: '简体中文',
      },
      {
        key: 'zhtw',
        name: 'zhtw',
        value: '繁體中文',
      },
      {
        key: 'th',
        name: 'th',
        value: 'ภาษาไทย',
      },
      {
        key: 'ja',
        name: 'ja',
        value: '日本語です',
      },
      {
        key: 'ko',
        name: 'ko',
        value: '한국어',
      },
      {
        key: 'tr',
        name: 'tr',
        value: 'Türkçe',
      },
      {
        key: 'my',
        name: 'my',
        value: 'မြန်မာဘာသာ',
      },
      {
        key: 'sv',
        name: 'sv',
        value: 'Svenska',
      },
    ],
    languageData: {
      en: '',
      zh: '',
      zhtw: '',
      th: '',
      ja: '',
      ko: '',
      tr: '',
      my: '',
      sv: '',
    },
    cloudBasic: {},
    isCloudDeploy: false,
    macData: {
      mac: '',
      uuid: '',
    },
    pageParams: {},
  }),
  getters: {
    currentLanguage: (state) => state.language,
    languageKeyMap: (state) =>
      state.languageList.reduce((acc, item) => {
        acc[item.key] = item;
        return acc;
      }, {}),
    languageNameMap: (state) =>
      state.languageList.reduce((acc, item) => {
        acc[item.name] = item;
        return acc;
      }, {}),
    languageValueMap: (state) =>
      state.languageList.reduce((acc, item) => {
        acc[item.value] = item;
        return acc;
      }, {}),

    languageKeys: (state) => state.languageList.map((item) => item.key),
    languageNames: (state) => state.languageList.map((item) => item.name),
    languageValues: (state) => state.languageList.map((item) => item.value),

    getLanguageByKey: (state) => {
      return (key) => state.languageList.find((item) => item.key === key);
    },
    getLanguageNameByKey: (state) => {
      return (key) => state.languageList.find((item) => item.key === key)?.name;
    },
    getLanguageValueByKey: (state) => {
      return (key) => state.languageList.find((item) => item.key === key)?.value;
    },

    getLanguageByName: (state) => {
      return (name) => state.languageList.find((item) => item.name === name);
    },
    getLanguageKeyByName: (state) => {
      return (name) => state.languageList.find((item) => item.name === name)?.key;
    },
    getLanguageValueByName: (state) => {
      return (name) => state.languageList.find((item) => item.name === name)?.value;
    },

    getLanguageByValue: (state) => {
      return (value) => state.languageList.find((item) => item.value === value);
    },
    getLanguageKeyByValue: (state) => {
      return (value) => state.languageList.find((item) => item.value === value)?.key;
    },
    getLanguageNameByValue: (state) => {
      return (value) => state.languageList.find((item) => item.value === value)?.name;
    },
  },
  actions: {
    setLanguage(lang) {
      if (lang) {
        this.language = lang;
      }
    },
    setLanguageList(data) {
      this.languageList = data;
    },
    setLanguageData(data) {
      this.languageData = data;
    },
    getLanguageData() {
      return {
        languageData: computed(() => {
          return this.languageData;
        }),
      };
    },
    setCloudBasic(data) {
      this.cloudBasic = data;
    },
    setMacData(data) {
      this.macData = data;
    },
    getMacData() {
      return {
        macData: computed(() => {
          return this.macData;
        }),
      };
    },

    setPageParams(data) {
      this.pageParams = data;
    },
    getPageParams() {
      return {
        pageParams: computed(() => {
          return this.pageParams;
        }),
      };
    },
    getCloudBasic() {
      return {
        cloudBasic: computed(() => {
          return this.cloudBasic;
        }),
      };
    },
    setIsCloudDeploy(data) {
      this.isCloudDeploy = data;
    },
    getIsCloudDeploy() {
      return {
        isCloudDeploy: computed(() => {
          return this.isCloudDeploy;
        }),
      };
    },
    getLanguageKey() {
      return {
        language: computed(() => {
          return this.language;
        }),
      };
    },
    getLanguage() {
      return {
        language: computed(() => {
          let result = 'English';
          (this.languageList || []).map((item) => {
            if (item?.name == this.language) {
              result = item.value;
            }
          });
          return result;
        }),
      };
    },
    getLanguageList() {
      return {
        languageList: computed(() => {
          return this.languageList;
        }),
      };
    },
    getLanguageListOrigin() {
      return {
        languageListOrigin: computed(() => {
          return this.languageListOrigin;
        }),
      };
    },
    getLanguageKeyForm() {
      return this.languageList.reduce((acc, item) => {
        acc[item.key] = '';
        return acc;
      }, {});
    },
    getLanguageNameForm() {
      return this.languageList.reduce((acc, item) => {
        acc[item.name] = '';
        return acc;
      }, {});
    },
    getLanguageValueForm() {
      return this.languageList.reduce((acc, item) => {
        acc[item.value] = '';
        return acc;
      }, {});
    },
  },
  persist: {
    key: 'Language',
    storage: localStorage,
  },
});
