import { JSEncrypt } from 'jsencrypt';
import CryptoJS from 'crypto-js';

import { message } from '@/utils/message.js';

/**
 * 加密
 * @param content 要加密的数据
 * @param publicKey 公钥
 * @return {Promise}
 */
export async function encrypt(content, publicKey) {
    try {
        // 假设你要加密的内容
        const symmetricKey = Array.from({length: 16}, () => Math.floor(Math.random() * 36).toString(36)).join('');
        const iv = Array.from({length: 16}, () => Math.floor(Math.random() * 36).toString(36)).join('');
        // 进行 AES 加密
        const ciphertext = CryptoJS.AES.encrypt(JSON.stringify(content), CryptoJS.enc.Utf8.parse(symmetricKey), {
            iv: CryptoJS.enc.Utf8.parse(iv),
            mode: CryptoJS.mode.CBC,
            padding: CryptoJS.pad.Pkcs7
        }).toString();
        // 使用 RSA 加密对称密钥
        const encrypt = new JSEncrypt();
        encrypt.setPublicKey(publicKey);
        const encryptedKey = encrypt.encrypt(symmetricKey);
        // 将加密后的数据、IV 和加密的对称密钥一起传输
        const finalData = btoa(CryptoJS.enc.Base64.stringify(CryptoJS.enc.Utf8.parse(iv)) + '||' + ciphertext + '||' + encryptedKey);
        // 
        return Promise.resolve(finalData);
    } catch (error) {
        message.error('encrypt => ' + $t('签名过期，请重试！'));
        return Promise.reject(error);
    }
}

/**
 * 解密
 * @param content 要解密的数据
 * @param privateKey 私钥
 * @param clientId 生成公钥、私钥的随机数
 * @returns {Promise}
 */
export async function decrypt(content, privateKey, clientId) {
    try {
        const contents = content.split('||');
        // 提取 IV、加密的数据和加密的对称密钥
        const encryptedKey = atob(contents[1]);
        // 用私钥解密…
        var decrypt = new JSEncrypt();
        decrypt.setPrivateKey(privateKey);
        const symmetricKey = decrypt.decrypt(btoa(encryptedKey));
        // 将 Base64 解码
        let encryptedDataBytes = CryptoJS.enc.Base64.parse(contents[0]);
        // 获取 IV 和密文
        let iv = CryptoJS.lib.WordArray.create(encryptedDataBytes.words.slice(0, 4)); // 16字节 IV
        let ciphertext = CryptoJS.lib.WordArray.create(encryptedDataBytes.words.slice(4)); // 剩余部分为密文，不取后面的256位数据
        // 使用相同的算法解密
        let decrypted = CryptoJS.AES.decrypt({ ciphertext: ciphertext }, CryptoJS.enc.Utf8.parse(symmetricKey), { iv: iv, mode: CryptoJS.mode.CBC, padding: CryptoJS.pad.Pkcs7 });
        // 转换回明文
        return Promise.resolve(JSON.parse(decrypted.toString(CryptoJS.enc.Utf8)));
    } catch (error) {
        message.error('decrypt => ' + $t('签名过期，请重试！'));
        return Promise.reject(error);
    }
}

/**
 * 创建密钥对
 * @param clientId 生成公钥、私钥的随机数·
 * @returns {Promise}
 */
export async function generateKeys(clientId) {
    try {
        var encrypt = new JSEncrypt();
        encrypt.setKey(encrypt.getPrivateKey(2048));
        const data = {
            publicKey: encrypt.getPublicKey(),
            privateKey: encrypt.getPrivateKey(),
            clientId
        };
        //
        return Promise.resolve(data);
    } catch (error) {
        message.error('Generate => ' + $t('签名过期，请重试！'));
        return Promise.reject(error);
    }
}
