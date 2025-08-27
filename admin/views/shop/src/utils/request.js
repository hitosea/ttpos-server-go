import axios from 'axios';
import qs from 'qs';
import router from '@/router';
import configObj from '@/config';
import { useUserStore } from '@/store';
import { message } from '@/utils/message.js';
import { useLockscreenStore } from '@/store/model/lockscreen';
import { encrypt, decrypt } from '@/utils/jsencrypt';
import { secretKeyStore } from '@/store/model/secretKey';

let { baseURL, tokenName, contentType, withCredentials, responseType } = configObj;
axios.defaults.headers['Content-Type'] = contentType; //配置请求头
axios.defaults.baseURL = baseURL;
axios.defaults.withCredentials = withCredentials;
axios.defaults.responseType = responseType;
axios.defaults.timeout = 60000; //请求超时
//POST传参序列化(添加请求拦截器)
axios.interceptors.request.use(
  async (config) => {
    // 是否加密的接口
    let encryptArr = ['shop/passport/login', '/shop/passport/saasEditPassword', '/shop/passport/editPass'];
    if (encryptArr.includes(config.url)) {
      config.encrypt = true;
    }
    //在发送请求之前做某件事
    const userStore = useUserStore();
    const { token, userInfo } = userStore;
    let lang = 'en';
    if (JSON.parse(localStorage.getItem('Language'))) {
      lang = JSON.parse(localStorage.getItem('Language')).language;
    }
    config.headers[tokenName] = token;
    config.headers['AppID'] = userInfo && userInfo.AppID;
    config.headers['language'] = lang;
    config.headers['version-name'] = import.meta.env.VITE_BASIC_VERSION;
    //浏览器缓存获取UUID+时间戳
    if (localStorage.getItem('uuid')) {
      config.headers['Suuid'] = localStorage.getItem('uuid') + '--' + new Date().getTime();
    }

    // 是否加密
    if (config.encrypt && config.data && ['post', 'put'].includes(config.method || '')) {
      const serverSecretKey = await secretKeyStore().getServerSecretKey();
      if (serverSecretKey.type === 'jsencrypt' && serverSecretKey.key) {
        // 加密数据
        const data = {};
        const encryptData = await encrypt(config.data, serverSecretKey.key);
        if (encryptData) data.encrypted = encryptData;
        config.data = data;
        config.headers['Content-Type'] = 'application/json';
        // 生成秘钥对，给后台传公钥
        const clientSecretKey = await secretKeyStore().getClientSecretKey();
        config.headers.encrypt = `encrypt_type=${serverSecretKey.type};encrypt_id=${serverSecretKey.id};client_type=${clientSecretKey.type};client_key=${clientSecretKey.publicKey}`;
      }
    } else if (config.method === 'post' && !config.headers.uploadImg) {
      config.data = qs.stringify(config.data);
    }
    return config;
  },
  (error) => {
    console.log('错误的传参');
    return Promise.reject(error);
  }
);

//返回状态判断(添加响应拦截器)
axios.interceptors.response.use(
  async (res) => {
    if (res.config.encrypt && ['post', 'put'].includes(res.config.method || '') && res.data.encrypted) {
      const clientSecretKey = await secretKeyStore().getClientSecretKey();
      res.data = await decrypt(res.data.encrypted, clientSecretKey.privateKey, clientSecretKey.id);
    }
    
    // 自动格式化返回数据中的数字，去掉多余的0
    if (res.data) {
      res.data = formatNumbers(res.data);
    }
    // 错误状态码并且不用弹窗的状态码
    let noErrorArr = [-901];
    //未登陆
    if (res.data.code !== 1) {
      if (res.data.code === 0) {
        !noErrorArr.includes(res?.data?.data?.code) && sendMessage(res.data.msg, 'error');
        return Promise.reject(res.data);
      } else if (res.data.code === -1) {
        sendMessage(res.data.msg, 'error');
        const userStore = useUserStore();
        const { afterLogout } = userStore;
        afterLogout();
        router.push({
          path: '/login',
        });
        return res.data;
      } else if (res.data.code === -101) {
        const useLockscreen = useLockscreenStore();
        useLockscreen.setAccredit(false);
        useLockscreen.setExpire(false);
        const userStore = useUserStore();
        const { afterLogout } = userStore;
        afterLogout();
        router.push({
          path: '/login',
        });
        return res.data;
      } else if (res.data.code === -102) {
        const useLockscreen = useLockscreenStore();
        useLockscreen.setExpire(true);
        return res.data;
      } else if (res.data.code === -201) {
        return res.data;
      } else if (res.data.code === 500) {
        return res.data;
      } else if (res.data.code) {
        const userStore = useUserStore();
        const { afterLogout } = userStore;
        afterLogout();
        router.push({
          path: '/login',
        });
      }
    } else {
      const useLockscreen = useLockscreenStore();
      useLockscreen.setAccredit(true);
      useLockscreen.setExpire(false);
      return res.data;
    }
  },
  (error) => {
    if (error.code == 'ECONNABORTED') {
      ElMessage;
      sendMessage($t('请求超时，请重试'), 'error');
    } else {
      ElMessage;
      sendMessage($t('接口请求异常，请稍后再试~'), 'error');
    }
    return Promise.reject(error);
  }
);

/**
 * 返回一个Promise(发送post请求)
 * errorback是否错误回调
 */
export function _post(url, params, errorback) {
  //有表情符号返回错误
  if (hasEmoji(JSON.stringify(params))) {
    message({
      message: $t('不能含有表情符号'),
      type: 'error',
    });
    return Promise.reject();
  }

  return new Promise((resolve, reject) => {
    axios
      .post(url, params)
      .then((response) => {
        resolve(response);
      })
      .catch((error) => {
        errorback && reject(error);
      });
  });
}

/**
 * 返回一个Promise(发送get请求)
 * errorback是否错误回调
 */
export function _get(url, param, errorback) {
  return new Promise((resolve, reject) => {
    axios
      .get(url, {
        params: param,
      })
      .then((response) => {
        resolve(response);
      })
      .catch((error) => {
        errorback && reject(error);
      });
  });
}
/**
 * 返回一个Promise(发送上传请求)
 * errorback是否错误回调
 */
export function _upload(url, formData, errorback) {
  return new Promise((resolve, reject) => {
    let headers = {
      'Content-Type': 'multipart/form-data',
      'uploadImg': true,
    };
    axios
      .post(url, formData, { headers })
      .then((response) => {
        resolve(response);
      })
      .catch((error) => {
        reject(error);
      });
  });
}

function sendMessage(e, type) {
  message({
    message: e,
    type: type,
  });
}

//去掉表情字符
function hasEmoji(input) {
  const emojiRegex = /[\uD800-\uDBFF][\uDC00-\uDFFF]|\uD83C[\uDF00-\uDFFF]|\uD83D[\uDC00-\uDE4F\uDE80-\uDEFF]/g;
  return emojiRegex.test(input);
}

// 递归格式化数字，去掉多余的0
function formatNumbers(data) {
  if (data === null || data === undefined) {
    return data;
  }

  // 如果是数组，递归处理每个元素
  if (Array.isArray(data)) {
    return data.map(item => formatNumbers(item));
  }

  // 如果是对象，递归处理每个属性
  if (typeof data === 'object') {
    const result = {};
    for (const key in data) {
      if (data.hasOwnProperty(key)) {
        result[key] = formatNumbers(data[key]);
      }
    }
    return result;
  }

  // 如果是数字类型，进行格式化
  if (typeof data === 'number') {
    // 将数字转换为字符串后再转回数字，自动去掉多余的0
    return parseFloat(data.toString());
  }

  // 如果是字符串且看起来像数字，也进行格式化
  if (typeof data === 'string' && /^\d+\.?\d*$/.test(data.trim())) {
    const num = parseFloat(data);
    // 如果转换成功且不是NaN，返回格式化后的数字
    if (!isNaN(num)) {
      return parseFloat(num.toString());
    }
  }

  // 其他类型保持不变
  return data;
}

export default {
  _post,
  _get,
  _upload,
};
