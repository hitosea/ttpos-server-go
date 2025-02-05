import { createApp } from 'vue';
import '../static/css/app.css';
import '../static/css/common.css';
// import router from './router';

import 'virtual:svg-icons-register';

import { loadDirectives } from '@/directive';
import filters from '@/filters/index.js';
import { setupRouter } from '@/router';
import * as ElementPlusIconsVue from '@element-plus/icons-vue';
import { createPinia } from 'pinia';
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate';
import VueClipboard from 'vue-clipboard2';
import VueUeditorWrap from 'vue-ueditor-wrap';
import App from './App.vue';
import I18n from './lang/index';
import { formatPrice, PriceCalculation, priceTwo } from './utils/formatPrice.js';
import { message } from './utils/message.js';
// import * as Sentry from '@sentry/vue';

const pinia = createPinia();
pinia.use(piniaPluginPersistedstate);

const app = createApp(App);

/** 加载自定义指令 */
loadDirectives(app);

// 注册 Element Plus 图标组件
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component);
}

// Sentry.init({
//   app,
//   dsn: import.meta.env.VITE_BASIC_DSN ?? 'https://1f027a65537c055bf6840c9b47218322@o4508408999051264.ingest.de.sentry.io/4508414999396432',
//   integrations: [Sentry.browserTracingIntegration(), Sentry.replayIntegration()],

//   // Set tracesSampleRate to 1.0 to capture 100%
//   // of transactions for tracing.
//   // We recommend adjusting this value in production
//   // Learn more at
//   // https://docs.sentry.io/platforms/javascript/configuration/options/#traces-sample-rate
//   tracesSampleRate: 1.0,

//   // Set `tracePropagationTargets` to control for which URLs trace propagation should be enabled
//   tracePropagationTargets: ['localhost', /^https:\/\/yourserver\.io\/api/],

//   release: 'TTPOS-Shop@1.0.9',
//   // Capture Replay for 10% of all sessions,
//   // plus for 100% of sessions with an error
//   // Learn more at
//   // https://docs.sentry.io/platforms/javascript/session-replay/configuration/#general-integration-configuration
//   replaysSessionSampleRate: 0.1,
//   replaysOnErrorSampleRate: 1.0,
// });

app.use(VueUeditorWrap);
app.use(VueClipboard);
app.use(I18n);
app.use(pinia);

// 设置路由
setupRouter(app);
// app.use(router);

app.mount('#app');

// 将$t对象添加到window中
window.$t = I18n.global.t;
// 将filters对象添加到全局属性中
app.config.globalProperties.$filter = filters;
// 将message对象添加到全局属性中
app.config.globalProperties.$ElMessage = message;
// 将priceTwo对象添加到全局属性中
app.config.globalProperties.$priceTwo = priceTwo;
// 将formatPrice对象添加到全局属性中
app.config.globalProperties.$formatPrice = formatPrice;
// 将PriceCalculation对象添加到全局属性中
app.config.globalProperties.$PriceCalculation = PriceCalculation;
