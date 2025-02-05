module.exports = {
  env: {
    browser: true,
    es2021: true,
    node: true,
  },
  globals: {},
  extends: ['eslint:recommended', 'plugin:@typescript-eslint/recommended', 'plugin:vue/vue3-essential', 'plugin:prettier/recommended', 'prettier'],
  overrides: [
    {
      env: {
        node: true,
      },
      files: ['.eslintrc.{js,cjs}'],
      parserOptions: {
        sourceType: 'script',
      },
    },
  ],
  parserOptions: {
    ecmaVersion: 'latest',
    parser: '@typescript-eslint/parser',
    sourceType: 'module',
  },
  plugins: ['@typescript-eslint', 'vue', 'import'],
  rules: {
    'prettier/prettier': 'error',
    // 关闭any类型判断
    '@typescript-eslint/no-explicit-any': 'off',
    // 关闭组件命名规则
    'vue/multi-word-component-names': 'off',
  },
};
