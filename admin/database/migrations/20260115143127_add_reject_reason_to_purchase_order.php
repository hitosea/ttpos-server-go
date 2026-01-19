<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddRejectReasonToPurchaseOrder extends Migrator
{
    /**
     * 为采购订单表新增驳回原因字段
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('purchase_order')) {
            $table = $this->table('purchase_order');
            
            // 检查字段是否不存在
            if (!$table->hasColumn('reject_reason')) {
                $table->addColumn('reject_reason', 'text', ['null' => true, 'comment' => '驳回原因', 'after' => 'reject_time']);
            }
            
            $table->update();
        }
    }
}



