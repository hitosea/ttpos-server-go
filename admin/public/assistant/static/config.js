window.config = () => {
    return {
        title: 'TTPOS', // 默认浏览器标题 JBCレジ TTPOS
        brand: 'ttpos', // 品牌 日本=> jbc, 泰国=>ttpos
        webLogo: '', // webLogo有值是优先加载Logo 否则按照地区加载默认图标
        mode: 'saas', // 部署模式 saas=云端部署 or local=局域网部署
        saasApiUrl: '', // saas部署的时候需要配置服务器地址，如果是局域网部署可不填写
        // eeui: true, //ios打包需要打开，勿删
    };
};
