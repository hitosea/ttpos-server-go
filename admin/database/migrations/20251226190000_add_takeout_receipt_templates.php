<?php

use think\facade\Db;
use think\facade\Log;
use think\migration\Migrator;
use think\migration\db\Column;

class AddTakeoutReceiptTemplates extends Migrator
{
    /**
     * 添加外卖商家联和顾客联模板
     * 
     * 此迁移文件会在 ttpos_printer_template 表中插入两条新记录：
     * - ID=12, UUID=12: 外卖商家联
     * - ID=13, UUID=13: 外卖顾客联
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);
        $nowTime = time();

        // 外卖票据模板列表
        $templates = [
            [
                'id' => 12,
                'uuid' => 13,
                'name' => '外卖商家联',
                'template' => 1,
                'is_show_sku' => 1,
                'tmp_uuid' => 0,
                'tmp_data' => '',
                'create_time' => $nowTime,
                'update_time' => $nowTime,
                'delete_time' => 0,
            ],
            [
                'id' => 13,
                'uuid' => 14,
                'name' => '外卖顾客联',
                'template' => 1,
                'is_show_sku' => 1,
                'tmp_uuid' => 0,
                'tmp_data' => '',
                'create_time' => $nowTime,
                'update_time' => $nowTime,
                'delete_time' => 0,
            ],
        ];

        try {
            foreach ($templates as $template) {
                // 检查记录是否已存在
                $exists = $db->table('ttpos_printer_template')
                    ->where('id', $template['id'])
                    ->find();

                if (!$exists) {
                    // 插入新记录
                    $db->table('ttpos_printer_template')->insert($template);
                } else {
                }
            }
            
        } catch (\Exception $e) {
            Log::error('外卖票据模板迁移失败: ' . $e->getMessage());
            throw $e;
        }
    }

}

