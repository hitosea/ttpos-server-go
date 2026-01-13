<?php

use think\facade\Db;
use think\migration\Migrator;

class AddPurchaseBrandLimitConfig extends Migrator
{
    /**
     * Change Method.
     *
     * 添加品牌采购限额全局配置到 ttpos_setting 表
     * 
     * 说明：
     * - purchase_brand_setting: 品牌采购设置，json格式，daily_limit: 每日申请次数上限，-1 表示不限制
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);

        // 检查配置是否已存在
        $existingConfigs = $db->name('setting')
            ->whereIn('key', ['purchase_brand_setting'])
            ->column('key');

        // 准备要插入的配置项
        $configs = [];
        
        if (!in_array('purchase_brand_setting', $existingConfigs)) {
            $configs[] = [
                'key' => 'purchase_brand_setting',
                'values' => json_encode(['daily_limit' => -1], JSON_UNESCAPED_UNICODE),
                'describe' => '品牌采购每日申请次数上限（-1表示不限制）',
                'create_time' => time(),
                'update_time' => time(),
            ];
        }

        // 批量插入
        if (!empty($configs)) {
            $db->name('setting')->insertAll($configs);
        }
    }
}

