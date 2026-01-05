<?php

use think\facade\Db;
use think\facade\Log;
use think\migration\Migrator;

class AddTakeoutRefundReceiptTemplate extends Migrator
{
    /**
     * 添加外卖退单联模板
     * 
     * 此迁移文件会在 ttpos_printer_template 表中插入一条新记录：
     * - ID=14, UUID=15: 外卖退单联
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);
        $nowTime = time();

        // 外卖退单联模板
        $template = [
            'id' => 14,
            'uuid' => 15,
            'name' => '外卖退单联',
            'template' => 1,
            'is_show_sku' => 1,
            'tmp_uuid' => 0,
            'tmp_data' => '',
            'create_time' => $nowTime,
            'update_time' => $nowTime,
            'delete_time' => 0,
        ];

        try {
            // 检查记录是否已存在
            $exists = $db->table('ttpos_printer_template')
                ->where('id', $template['id'])
                ->find();

            if (!$exists) {
                // 插入新记录
                $db->table('ttpos_printer_template')->insert($template);
                Log::info('外卖退单联模板添加成功: ID=' . $template['id'] . ', UUID=' . $template['uuid']);
            } else {
                Log::info('外卖退单联模板已存在，跳过插入');
            }
            
        } catch (\Exception $e) {
            Log::error('外卖退单联模板迁移失败: ' . $e->getMessage());
            throw $e;
        }
    }
}

