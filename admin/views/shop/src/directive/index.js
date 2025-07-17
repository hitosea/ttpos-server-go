import {
    defaultImg
} from '../config/env.js'
import {
    getSessionStorage,
    setSessionStorage
} from '@/utils/base.js'
import {
    getStorage
} from '@/utils/storageData';
import {
    createdAuth
} from '@/utils/createdAuth.js'
import configObj from "@/config";
let { menu } = configObj;
/** 挂载自定义指令 */
export function loadDirectives(app) {
    /*指令测试*/
    app.directive('demo', {
        bind: function (el, binding, vnode) {
            var s = JSON.stringify
            el.innerHTML =
                'name: ' + s(binding.name) + '<br>' +
                'value: ' + s(binding.value) + '<br>' +
                'expression: ' + s(binding.expression) + '<br>' +
                'argument: ' + s(binding.arg) + '<br>' +
                'modifiers: ' + s(binding.modifiers) + '<br>' +
                'vnode keys: ' + Object.keys(vnode).join(', ')
        }
    })

    /*权限*/
    app.directive('auth', {
        mounted(el, binding, vnode) {
            // console.log(binding.value)
            let auth = getSessionStorage('authlist');
            if (!(auth && JSON.stringify(auth) !== "{}")) {
                let authlist = {}
                auth = getStorage(menu);
                createdAuth(auth, authlist);
                setSessionStorage('authlist', authlist);
                auth = authlist;
            }
            let value = binding.value.toLowerCase();
            if (auth[value] != true) {
                el.style.display = "none";
            }
        }
    })

    /*默认图片*/
    app.directive('img-url', async function (el, binding) {
        let imgURL = '';
        if (binding.value instanceof Object) {
            let jsonStr = binding.expression.split(',')[0].replace(/\'/g, '');
            imgURL = checkChild(binding.value, jsonStr);
        } else {
            imgURL = binding.value;
        }

        if (imgURL && !imgURL.includes('undefined') && imgURL != '/') {
            try {
                let exist = await imageIsExist(imgURL);
                if (exist) {
                    el.setAttribute('src', imgURL);
                } else {
                    el.setAttribute('src', defaultImg);
                }
            } catch (error) {
                console.error('Image check failed:', error);
                el.setAttribute('src', defaultImg);
            }
        } else {
            el.setAttribute('src', defaultImg);
        }
    });
}





/*检查子对象是否为NUll*/
let checkChild = function (obj, str) {
    let _img = null;
    let arr = str.match(/\./g);
    if (!arr) {
        _img = obj[str];
    } else {
        let _i = str.indexOf('.');
        let key = str.substring(0, _i);
        let newobj = obj[key];
        if (newobj instanceof Object == true) {
            let newstr = str.substring(_i + 1);
            _img = checkChild(newobj, newstr);
        }
    }
    return _img;
}

/* 检测图片是否存在*/
async function imageIsExist(url, timeout = 30000) {
    return new Promise((resolve, reject) => {
        // 统一使用Image对象，避免fetch的跨域问题
        const img = new Image();
        let isResolved = false;
        
        const cleanup = () => {
            img.onload = img.onerror = null;
            if (timer) clearTimeout(timer);
        };
        
        const timer = setTimeout(() => {
            if (isResolved) return;
            isResolved = true;
            cleanup();
            resolve(false); // 超时返回false而不是reject
        }, timeout);

        img.onload = () => {
            if (isResolved) return;
            isResolved = true;
            cleanup();
            resolve(true);
        };

        // 第一步：智能判断是否需要CORS
        if (isGoogleCloudStorageUrl(url)) {
            img.src = url; // 直接加载，不设置crossOrigin
        } else {
            img.crossOrigin = 'anonymous'; // 其他URL尝试CORS
            img.src = url;
        }

        // 第二步：如果CORS失败，自动降级
        img.onerror = () => {
            if (img.crossOrigin) {
                // 重新尝试不设置CORS
                imageIsExistWithoutCORS(url, timeout).then(resolve);
            }
        };
    });
}

/* 不设置CORS的图片检测（备用方案）*/
function imageIsExistWithoutCORS(url, timeout = 30000) {
    return new Promise((resolve) => {
        const img = new Image();
        let isResolved = false;
        
        const cleanup = () => {
            img.onload = img.onerror = null;
            if (timer) clearTimeout(timer);
        };
        
        const timer = setTimeout(() => {
            if (isResolved) return;
            isResolved = true;
            cleanup();
            resolve(false);
        }, timeout);

        img.onload = () => {
            if (isResolved) return;
            isResolved = true;
            cleanup();
            resolve(true);
        };

        img.onerror = () => {
            if (isResolved) return;
            isResolved = true;
            cleanup();
            resolve(false);
        };

        // 不设置crossOrigin，避免CORS问题
        img.src = url;
    });
}

/* 检查是否为Google Cloud Storage URL */
function isGoogleCloudStorageUrl(url) {
    return url.includes('storage_googleapis') || 
           url.includes('storage.googleapis.com') ||
           url.includes('GoogleAccessId=');
}

/* 检查是否为已知的CORS友好域名 */
function isKnownCORSFriendlyDomain(url) {
    try {
        const urlObj = new URL(url);
        const hostname = urlObj.hostname.toLowerCase();
        
        // 已知不支持CORS的域名列表
        const nonCORSDomains = [
            'ttpos-test1.ttpos.com', // 你的域名
            // 可以根据需要添加更多
        ];
        
        return nonCORSDomains.some(domain => hostname.includes(domain));
    } catch {
        return false;
    }
}