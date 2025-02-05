import { defineConfig, loadEnv } from 'vite';
import { resolve } from 'path';
import vue from '@vitejs/plugin-vue';

import AutoImport from 'unplugin-auto-import/vite';
import Components from 'unplugin-vue-components/vite';
import vitePluginCompression from 'vite-plugin-compression';

import { ElementPlusResolver } from 'unplugin-vue-components/resolvers';
import { createSvgIconsPlugin } from 'vite-plugin-svg-icons';
import { createStyleImportPlugin, ElementPlusResolve } from 'vite-plugin-style-import';

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  return {
    resolve: {
      // 配置路径别名
      alias: {
        '@/': `${resolve(process.cwd(), './src')}/`,
      },
    },
    plugins: [
      vue(),
      AutoImport({
        resolvers: [ElementPlusResolver()],
      }),
      Components({
        resolvers: [ElementPlusResolver()],
      }),
      createStyleImportPlugin({
        resolves: [ElementPlusResolve()],
      }),
      createSvgIconsPlugin({
        // 指定要缓存的文件夹
        iconDirs: [resolve(process.cwd(), './src/assets/icons')],
        // 指定symbolId格式
        symbolId: 'icon-[name]',
      }),
      vitePluginCompression({
        threshold: 1024 * 1024 * 1, // 对大于 1MB 的文件进行压缩
        // deleteOriginFile: true,
      }),
    ],
    server: {
      host: '0.0.0.0', // 服务器ip地址
      port: +env.VITE_APP_PORT || 3000, // 本地端口
      // open: true, // 是否自动在浏览器打开
      proxy: {
        '/api/': {
          target: env.VITE_API_URL,
          changeOrigin: true,
          // secure: false, // 验证 SSL 证书
          // rewrite: (path) => path.replace(/^\/api/, '/api/v1/'), // 重写传过来的path路径
        },
      },
    },
    // root: process.cwd(), // 项目根目录
    base: env.NODE_ENV == 'development' ? '' : '/admin/', // 公共基础路径
    build: {
      outDir: 'admin', // 输出目录（部署需要这个名称，勿删）
      // assetsDir: 'assets', // 指定生成静态资源的存放路径
      // cssTarget: 'chrome80', // css的编译目标
      target: 'es2015', // 设置最终构建的浏览器兼容目标。默认值是一个 Vite 特有的值——'modules'  还可设置为 'es2015' 'es2016'等
      cssCodeSplit: true, // 如果设置为false，整个项目中的所有 CSS 将被提取到一个 CSS 文件中
      manifest: false, // 是否产出manifest.json, 用于打包分析
      sourcemap: false, // 构建后是否生成 source map 文件。如果为 true，将会创建一个独立的 source map 文件
      chunkSizeWarningLimit: 1000, // 单位kb  打包后文件大小警告的限制 (文件大于此此值会出现警告)
      assetsInlineLimit: 4096, // 单位字节（1024等于1kb） 小于此阈值的导入或引用资源将内联为 base64 编码，以避免额外的 http 请求。设置为 0 可以完全禁用此项。
      minify: 'terser', // 'terser' 相对较慢，但大多数情况下构建后的文件体积更小。'esbuild' 最小化混淆更快但构建后的文件相对更大。
      terserOptions: {
        compress: {
          drop_console: true, // 生产环境去除console
          drop_debugger: true, // 生产环境去除debugger
        },
      },
      rollupOptions: {
        output: {
          chunkFileNames: 'assets/js/[name]-[hash].js',
          entryFileNames: 'assets/js/[name]-[hash].js',
          assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
          manualChunks: {
            // 分包规则
            vue: ['vue', 'vue-i18n', 'vue-router', 'pinia', 'vue-query', 'axios'], // vue
            element: ['element-plus'], // UI库
            // openpgp: ['openpgp', 'localforage'], // 加密库
          },
        },
      },
    },
  };
});
