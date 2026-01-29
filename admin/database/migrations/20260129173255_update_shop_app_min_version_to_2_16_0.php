<?php

use think\migration\Migrator;

class UpdateShopAppMinVersionTo2160 extends Migrator
{
    const TARGET = 'main';

    /**
     * Change Method.
     *
     * 更新 shop_app 设置的最小版本号为 2.16.0
     * 如果记录不存在则新增
     */
    public function change()
    {
        $values = json_encode(['min_version' => '2.16.0'], JSON_UNESCAPED_UNICODE);
        $existing = $this->fetchRow("SELECT `key` FROM `ttpos_setting` WHERE `key` = 'shop_app'");

        if ($existing) {
            $this->execute("UPDATE `ttpos_setting` SET `values` = '{$values}', `update_time` = " . time() . " WHERE `key` = 'shop_app'");
        } else {
            $this->execute("INSERT INTO `ttpos_setting` (`key`, `describe`, `values`, `create_time`, `update_time`) VALUES ('shop_app', 'Shop端应用配置', '{$values}', " . time() . ", " . time() . ")");
        }
    }
}
