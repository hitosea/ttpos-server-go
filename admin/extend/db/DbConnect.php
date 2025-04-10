<?php
namespace db;

use think\facade\Db;
use think\facade\Config;

class DbConnect 
{
	/**
     * 连接数据库
     */
    public static function switch($appId=0, bool $forceDefault = true) {
        $dbName = $appId ? ('shop' . $appId) : env('DB_DATABASE');
        $connection = $appId ? $dbName : 'mysql';
        // 
        $config = config('database');
        $mysql = $config['connections']['mysql'];
        $mysql['database'] = $dbName;
        $mysql['username'] = env('DB_USERNAME');
        $mysql['password'] = env('DB_ROOT_PASSWORD');
        $config['connections'] = array_merge($config['connections'], [$dbName => $mysql]);
        if ($forceDefault) {
            $config['default'] = $connection;
        }
        Config::set($config, 'database');
        if ($forceDefault) {
            Db::connect($connection, false);
        }
        // 
        return $connection;
    }

}
