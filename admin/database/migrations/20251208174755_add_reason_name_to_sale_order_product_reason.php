<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddReasonNameToSaleOrderProductReason extends Migrator
{
    /**
     * 为销售订单产品原因表新增原因名称快照字段
     * Requirement: story-main-reason-snapshot-fix
     * Purpose: 保存免单/退菜时的原因名称快照（JSON），不随后台配置变更而改变
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('sale_order_product_reason')) {
            $table = $this->table('sale_order_product_reason');

            // 检查字段是否不存在（幂等性）
            if (!$table->hasColumn('name')) {
                $table->addColumn(
                    'name',
                    'text',
                    [
                        'comment' => '原因名称快照（JSON），不随后台更新',
                        'after' => 'gift_reason_uuid'
                    ]
                );

                $table->update();
            }
        }
    }
}

