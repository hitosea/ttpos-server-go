<?php
/**
 * 为采购单表的 final_receive_time 字段添加索引
 *
 * 查询：GetLastCompletedBrandPurchaseInfo 使用 final_receive_time 做范围过滤和 JOIN 匹配
 * 原因：final_receive_time 字段无索引，WHERE/JOIN 条件会全表扫描
 */

use think\migration\Migrator;

class AddIndexFinalReceiveTimeToPurchaseOrder extends Migrator
{
    public function change()
    {
        if ($this->hasTable('purchase_order')) {
            $table = $this->table('purchase_order');

            if ($table->hasColumn('final_receive_time') && !$table->hasIndex('final_receive_time')) {
                $table->addIndex('final_receive_time', ['name' => 'idx_final_receive_time']);
                $table->update();
            }
        }
    }
}
