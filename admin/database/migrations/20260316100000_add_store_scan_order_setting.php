<?php

use think\migration\Migrator;

class AddStoreScanOrderSetting extends Migrator
{
    /**
     * Change Method.
     *
     * 为已有商户添加门店点餐默认设置到 ttpos_setting 表
     * is_enabled=1, enable_delivery=1, enable_self_pickup=1
     */
    public function change()
    {
        $existing = $this->fetchRow("SELECT `key` FROM `ttpos_setting` WHERE `key` = 'store_scan_order'");
        if (!$existing) {
            $values = json_encode(['is_enabled' => 1, 'enable_delivery' => 1, 'enable_self_pickup' => 1], JSON_UNESCAPED_UNICODE);
            $this->execute("INSERT INTO `ttpos_setting` (`key`, `describe`, `values`, `create_time`, `update_time`) VALUES ('store_scan_order', '门店点餐设置', '{$values}', " . time() . ", " . time() . ")");
        }
    }
}
