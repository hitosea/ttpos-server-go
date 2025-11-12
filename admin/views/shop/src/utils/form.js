import IndexApi from '@/api/index.js';

/**
 * 生成唯一名称验证器函数
 * @param {string} source - 代表业务逻辑
 * @param {string|number}? id - 代表编辑id
 * @param {string}? lang - 多语言标识
 * @param {object}? message - 错误信息
 * @param {string}? message.empty - 空值错误信息
 * @param {string}? message.exist - 已存在错误信息
 * @returns {Function} 返回验证器函数
 */

export const uniqueNameValidator = (source, id, lang, message, required = true) => {
  let quiver = null;
  const messageEmpty = message?.empty ?? window.$t('请输入名称');
  //   const messageExist = message?.exist ?? window.$t('此名称已存在');
  return (rule, value, callback) => {
    if (!value) {
      if (required) {
        callback(messageEmpty);
      } else {
        callback();
      }
    } else {
      clearTimeout(quiver);
      callback();
      // 2025-11-12 14:35:56 去掉名称唯一
      // quiver = setTimeout(async () => {
      //     const params = {
      //         name: { [`${lang}`]: value },
      //         source,
      //         id,
      //     };
      //     try {
      //         const { data } = await IndexApi.checkNameExist(params, true);
      //         if (data[`${lang}`]) {
      //             callback(messageExist);
      //         } else {
      //             callback();
      //         }
      //     } catch (error) {
      //         console.log('uniqueNameValidator error', error);
      //         if (hasEmoji(value)) {
      //             callback(window.$t('不能含有表情符号'));
      //         } else {
      //             callback(error?.message ?? window.$t('网络请求错误'));
      //         }
      //     }
      // }, 300);
    }
  };
};
//去掉表情字符
function hasEmoji(input) {
  const emojiRegex = /[\uD800-\uDBFF][\uDC00-\uDFFF]|\uD83C[\uDF00-\uDFFF]|\uD83D[\uDC00-\uDE4F\uDE80-\uDEFF]/g;
  return emojiRegex.test(input);
}
