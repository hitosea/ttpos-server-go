<?php

use think\migration\Migrator;

class AddShopAppSetting extends Migrator
{
    const TARGET = 'main';

    /**
     * Change Method.
     *
     * 添加 shop_app 设置到 ttpos_setting 表
     * 用于配置 shop 端应用的最小版本号要求
     */
    public function change()
    {
        // 检查并插入 shop_app 设置（只有当key不存在时才插入）
        $existing = $this->fetchRow("SELECT `key` FROM `ttpos_setting` WHERE `key` = 'shop_app'");
        if (!$existing) {
            $values = json_encode(['min_version' => '2.15.0'], JSON_UNESCAPED_UNICODE);
            $this->execute("INSERT INTO `ttpos_setting` (`key`, `describe`, `values`, `create_time`, `update_time`) VALUES ('shop_app', 'Shop端应用配置', '{$values}', " . time() . ", " . time() . ")");
        }
    }
}
