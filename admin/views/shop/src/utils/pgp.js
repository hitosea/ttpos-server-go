import { encrypt, readKey, createMessage, decrypt, readMessage, decryptKey, readPrivateKey, generateKey } from 'openpgp/lightweight';

import { message } from '@/utils/message.js';

/**
 * 加密
 * @param content 要加密的数据
 * @param publicKey 公钥
 * @return {Promise}
 */
export async function pgpEncrypt(content, publicKey) {
    try {
        const encrypted = await encrypt({
            message: await createMessage({ text: JSON.stringify(content) }),
            encryptionKeys: await readKey({ armoredKey: publicKey }),
            date: new Date(Date.now() + 60 * 1000), // 延迟一分钟
        });
        return Promise.resolve(encrypted.toString().replace(/\s*-----(BEGIN|END) PGP MESSAGE-----\s*/g, ''));
    } catch (error) {
        message.error('pgpEncrypt => ' + $t('签名过期，请重试！'));
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
export async function pgpDecrypt(content, privateKey, clientId) {
    try {
        const { data } = await decrypt({
            message: await readMessage({
                armoredMessage: '-----BEGIN PGP MESSAGE-----\n\n' + content + '\n-----END PGP MESSAGE-----',
            }),
            decryptionKeys: await decryptKey({
                privateKey: await readPrivateKey({ armoredKey: privateKey }),
                passphrase: clientId,
            }),
        });
        return Promise.resolve(JSON.parse(data.toString()));
    } catch (error) {
        message.error('pgpDecrypt => ' + $t('签名过期，请重试！'));
        return Promise.reject(error);
    }
}

/**
 * 创建密钥对
 * @param clientId 生成公钥、私钥的随机数
 * @returns {Promise}
 */
export async function pgpGenerate(clientId) {
    try {
        const data = await generateKey({
            type: 'ecc',
            curve: 'curve25519',
            passphrase: clientId,
            userIDs: [{ name: 'doo', email: 'admin@admin.com' }],
        });
        //
        return Promise.resolve(data);
    } catch (error) {
        message.error('pgpGenerate => ' + $t('签名过期，请重试！'));
        return Promise.reject(error);
    }
}
