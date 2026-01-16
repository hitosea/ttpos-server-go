<?php

use think\migration\Migrator;

/**
 * 为采购订单表添加驳回原因字段
 * 
 * 在 ttpos_purchase_order 表中添加 reject_reason 字段，用于记录订单被驳回时的原因说明
 */
class AddRejectReasonToPurchaseOrder extends Migrator
{
    /**
     * 执行迁移
     */
    public function up()
    {
        // 检查表是否存在
        if (!$this->hasTable('purchase_order')) {
            return;
        }

        // 获取表对象
        $table = $this->table('purchase_order');
        
        // 检查字段是否已存在
        if ($table->hasColumn('reject_reason')) {
            return;
        }

        // 添加 reject_reason 字段
        // 添加到 reject_time 字段之后
        $this->execute("ALTER TABLE `ttpos_purchase_order` ADD COLUMN `reject_reason` TEXT NULL COMMENT '驳回原因' AFTER `reject_time`");
    }

    /**
     * 回滚迁移
     */
    public function down()
    {
        // 检查表是否存在
        if (!$this->hasTable('purchase_order')) {
            return;
        }

        // 获取表对象
        $table = $this->table('purchase_order');
        
        // 检查字段是否存在
        if (!$table->hasColumn('reject_reason')) {
            return;
        }

        // 删除字段
        $table->removeColumn('reject_reason')
              ->update();
    }
}
