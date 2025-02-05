import { createRouter, createWebHashHistory } from 'vue-router';
import routes from './routes';
// 实例化路由
const router = createRouter({ history: createWebHashHistory(), routes });
//
export { router };
