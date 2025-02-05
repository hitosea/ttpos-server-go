/** @type {import('tailwindcss').Config} */
export default {
  corePlugins: {
    preflight: false, // 禁用 Tailwind 的全局基本样式
  },
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: 'var(--el-color-primary)',
      },
    },
  },
  plugins: [],
};
