import 'nprogress/nprogress.css';
import NProgress from 'nprogress';
//
import { router } from './index';
import { constantRoutes } from './routes';
import { getToken } from '@/utils/auth';
const title = JSON.parse(localStorage.getItem('TITLE_BASE') || '{}').browser_title || '';
// NProgress Configuration
NProgress.configure({ easing: 'ease', speed: 500, showSpinner: false });
//
router.beforeEach(async (to, _, next) => {
  // 打开进度条
  NProgress.start();
  // 设置title
  if (title) {
    document.title = (to.meta?.title ? `${to.meta?.title} - ` : '') + title;
  } else {
    document.title = to.meta?.title ? `${to.meta?.title}` : '';
  }
  // token存在
  if (getToken()) {
    // 如果已经登录，不可进入登录页面
    if (['/login'].includes(to.path as any)) return next(to.query.redirect?.toString() ?? '/');
    // 跳转
    return next();
  }
  // 根据白名单的登录状态进行鉴权
  if (constantRoutes.findIndex((item) => item.path === to.path) > -1) return next();
  // 跳转登录
  return next('/login');
});

//
router.afterEach(() => {
  // 关闭进度条
  NProgress.done();
});

//
router.onError(() => {
  NProgress.done(true);
});
