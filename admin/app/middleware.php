<?php
// 全局中间件定义文件
return [
    // 全局请求缓存
    // \think\middleware\CheckRequestCache::class,
    // 多语言加载
    // \think\middleware\LoadLangPack::class,
    // Session初始化
    // \think\middleware\SessionInit::class,
    // 允许跨域请求
    app\common\middleware\AllowOrigin::class,
    // 授权
    // app\common\middleware\License::class,
    //安全验证
    app\common\middleware\ChenkRequest::class,
    // 删除缓存
    // app\shop\middleware\DeleteCache::class
];
