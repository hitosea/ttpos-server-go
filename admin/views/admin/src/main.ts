import { createApp } from 'vue';
import { VueQueryPlugin } from 'vue-query';
import { router } from '@/router';
import { store } from '@/stores';
import { i18n } from '@/i18n';
import { VPermission } from '@/directives/v-permission';
//
import './assets/css/index.scss';
import 'virtual:svg-icons-register';
//
import ElementPlus from 'element-plus';
import Icons from '@/plugins/element';
import App from './App.vue';
// 路由守卫，中间件
import '@/router/middleware';
// 检查版本自动更新
// import '@/utils/auto-update';
// import * as Sentry from '@sentry/vue';

// 实例化对象
const app = createApp(App);

// Sentry.init({
//   app,
//   dsn: import.meta.env.VITE_BASIC_DSN ?? 'ttps://de5e5d3f4a16a528b5fc280032a2b5ca@o4508408999051264.ingest.de.sentry.io/4508415000576080',
//   integrations: [Sentry.browserTracingIntegration(), Sentry.replayIntegration()],

//   // Set tracesSampleRate to 1.0 to capture 100%
//   // of transactions for tracing.
//   // We recommend adjusting this value in production
//   // Learn more at
//   // https://docs.sentry.io/platforms/javascript/configuration/options/#traces-sample-rate
//   tracesSampleRate: 1.0,

//   // Set `tracePropagationTargets` to control for which URLs trace propagation should be enabled
//   tracePropagationTargets: ['localhost', /^https:\/\/yourserver\.io\/api/],
//   release: 'TTPOS-Admin@1.0.9',
//   // Capture Replay for 10% of all sessions,
//   // plus for 100% of sessions with an error
//   // Learn more at
//   // https://docs.sentry.io/platforms/javascript/session-replay/configuration/#general-integration-configuration
//   replaysSessionSampleRate: 0.1,
//   replaysOnErrorSampleRate: 1.0,
// });

// i18n翻译
app.use(i18n);
// 状态管理
app.use(store);
// 路由
app.use(router);
//
app.use(ElementPlus);
// 全局注册ElementPlus图标
app.use(Icons);
//
app.use(VueQueryPlugin, {
  queryClientConfig: {
    defaultOptions: {
      queries: {
        refetchOnWindowFocus: false, // 窗口焦点影响的数据刷新
        retry: false, // 出错后不重试
      },
    },
  },
});
// 注册权限指令
app.directive('permission', VPermission);
// 挂载
app.mount('#app');
