import localForage from 'localforage';
import { defineStore } from 'pinia';
import { reactive, ref } from 'vue';

import { getPublicKey } from '@/api/login';
import { randomString } from '@/utils';
import { generateKeys } from '@/utils/jsencrypt';

export const secretKeyStore = defineStore('secretKeyStore', () => {
  // 客户端随机ID
  const clientId = ref();
  // 客户端公钥、私钥，加密类型
  const clientSecretKey = reactive<{
    type: string;
    id: string;
    publicKey: string;
    privateKey: string;
  }>({
    type: 'jsencrypt',
    id: '',
    publicKey: '',
    privateKey: '',
  });
  // 服务端公钥，加密类型
  const serverSecretKey = reactive<{ type: string; id: string; key: string }>({
    type: '',
    id: '',
    key: '',
  });

  //
  const getClientId = async () => {
    if (clientId.value) return clientId.value;
    // 本地数据库
    clientId.value = await localForage?.getItem('clientId');
    // 如果没有随机数，重新生成
    if (!clientId.value) {
      clientId.value = randomString();
      await localForage.setItem('clientId', clientId.value);
    }
    //
    return Promise.resolve(clientId.value);
  };
  //
  const getClientSecretKey = async () => {
    if (clientSecretKey.id && clientSecretKey.publicKey && clientSecretKey.privateKey) return clientSecretKey;
    const clientId = await getClientId();
    const res = await generateKeys(clientId);
    clientSecretKey.id = clientId;
    clientSecretKey.privateKey = res.privateKey;
    clientSecretKey.publicKey = res.publicKey
      .replace(/\s*-----(BEGIN|END) PUBLIC KEY-----\s*/g, '')
      .replace(/\n+/g, '$')
      .replace(/\+/g, '-')
      .replace(/\//g, '_');
    //
    return Promise.resolve(clientSecretKey);
  };
  //
  const getServerSecretKey = async () => {
    if (serverSecretKey.id && serverSecretKey.key && serverSecretKey.type) return serverSecretKey;
    // 如果不存在就重新请求接口
    try {
      const clientId = await getClientId();
      const res = await getPublicKey({ client_id: clientId, type: 'jsencrypt' });
      //
      serverSecretKey.id = res.data?.id;
      serverSecretKey.key = res.data?.key;
      serverSecretKey.type = res.data?.type;
      //
      return Promise.resolve(serverSecretKey);
    } catch (error) {
      return Promise.reject(error);
    }
  };
  return {
    getClientId,
    getClientSecretKey,
    getServerSecretKey,
  };
});
