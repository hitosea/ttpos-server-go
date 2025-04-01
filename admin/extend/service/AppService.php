<?php
declare (strict_types = 1);

namespace service;

use think\Service;
use help\ValidateHelp;
use think\facade\Config;
use think\facade\Db;
use think\facade\Validate;

/**
 * 应用服务类
 */
class AppService extends Service
{
    public function register()
    {
        // 服务注册
        
        // 开发模式 - 自动检测更新
        if (env('APP_DEBUG') == true) {
            ini_set('opcache.validate_timestamps', 1);
        }
    }

    public function boot()
    {
        // apidoc 数据库配置
        if (strpos(request()->url(), '/apidoc/') !== false) {
            $name = 'shop' . Db::name('company')->where('delete_time', 0)->value('uuid');
            $config = config('database');
            if (!isset($config['connections'][$name])) {
                $mysql = $config['connections']['mysql'];
                $mysql['database'] = $name;
                $mysql['username'] = 'root';
                $mysql['password'] = env('DB_ROOT_PASSWORD');
                $config['connections']['mysql'] = $mysql;
                Config::set($config, 'database');
                Db::connect('mysql', true);
            }
        }

        // 覆盖 MigrateRun 指令
        if ($this->app->runningInConsole()) {
            $selectPath = root_path('vendor/topthink/think-migration/src').'Service.php';
            $content = file_get_contents($selectPath);
            if (strpos($content, 'MigrateRun::class,') !== false && strpos($content, '//MigrateRun::class') === false) {
                $content = str_replace('MigrateRun::class,', '//MigrateRun::class,', $content);
                $content = str_replace('MigrateRollback::class,', '//MigrateRollback::class,', $content);
                file_put_contents($selectPath, $content);
            }
        }

        // 覆盖 Cache 指令
        if (!strstr(file_get_contents(root_path('vendor/topthink/framework/src/think').'Cache.php'), '@ver-0.0.2')) {
            $selectPath = root_path('extend/cache').'Cache.php';
            file_put_contents(root_path('vendor/topthink/framework/src/think').'Cache.php', file_get_contents($selectPath));
        }

        // 自定义验证规则
        Validate::maker(function($validate) {

            // 验证用户名
            $validate->extend('checkUserName', function ($value) {
                return preg_match('/^[a-zA-Z0-9]+$/', $value) ? true : false;
            }, '用户名必须是英文字符、数字' );

            // 验证密码
            $validate->extend('checkPassword', function ($value) {
                return ValidateHelp::validateAlphaPassword($value);
            }, '不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种' );

            // 验证数组整数
            $validate->extend('integerArray', function ($value) {
                if (!is_array($value)) {
                    return false;
                }
                foreach ($value as $item) {
                    if (!is_int($item)) {
                        return false;
                    }
                }
                return true;
            }, ':attribute数值必须为整数' );
        });
    }
}
