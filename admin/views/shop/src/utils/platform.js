export function ISEEUI() {
    var userAgent = navigator.userAgent;
    return userAgent.includes('android_kuaifan_eeui');
}

export function EEUIRELOAD() {
    if (ISEEUI()) {
        let url = window.location.href;
        let key = '_=';
        let reg = new RegExp(key + '\\d+');
        let timestamp = Math.round(new Date().getTime() / 1000);
        if (url.indexOf(key) > -1) {
            url = url.replace(reg, key + timestamp);
        } else {
            if (url.indexOf('\?') > -1) {
                let urlArr = url.split('\?');
                if (urlArr[1]) {
                    url = urlArr[0] + '?' + key + timestamp + '&' + urlArr[1];
                } else {
                    url = urlArr[0] + '?' + key + timestamp;
                }
            } else {
                if (url.indexOf('#') > -1) {
                    url = url.split('#')[0] + '?' + key + timestamp + location.hash;
                } else {
                    url = url + '?' + key + timestamp;
                }
            }
        }
        let webview = requireModuleJs('webview');
        webview.setUrl(url);
        // requireModuleJs("webview").setUrl(url);
        // requireModuleJs("webview").setUrl(url)
    } else {
        location.reload();
    }
}