<?php

use think\migration\Migrator;

/**
 * 更新采购订单状态字段注释
 * 
 * 将 ttpos_purchase_order 表的 status 字段注释更新为最新的状态定义
 */
class UpdatePurchaseOrderStatusComment extends Migrator
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

        // 获取表对象（只查询一次表结构）
        $table = $this->table('purchase_order');
        
        // 检查字段是否存在
        if (!$table->hasColumn('status')) {
            return;
        }

        // 更新 status 字段注释
        $this->execute("ALTER TABLE `ttpos_purchase_order` MODIFY COLUMN `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态: 0-草稿(待提交) 1-待审核 2-已审核(采购中/待收货) 3-已驳回 4-已完成(全部收货/已收货) 5-待总部审核'");
    }

}
