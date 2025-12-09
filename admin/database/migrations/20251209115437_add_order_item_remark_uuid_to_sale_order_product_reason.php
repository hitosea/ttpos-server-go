<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddOrderItemRemarkUuidToSaleOrderProductReason extends Migrator
{
    /**
     * 为销售订单产品原因表新增备注预设UUID字段
     * Requirement: story-order-item-remark-preset
     * Purpose: 支持在订单商品上关联备注预设，扩展 ttpos_sale_order_product_reason 表的用途
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('sale_order_product_reason')) {
            $table = $this->table('sale_order_product_reason');

            // 检查字段是否不存在（幂等性）
            if (!$table->hasColumn('order_item_remark_uuid')) {
                $table->addColumn(
                    'order_item_remark_uuid',
                    'biginteger',
                    [
                        'signed' => false,
                        'default' => 0,
                        'comment' => '备注预设UUID',
                        'after' => 'gift_reason_uuid'
                    ]
                );

                $table->update();
            }

            // 检查索引是否不存在（幂等性）
            if (!$table->hasIndex(['order_item_remark_uuid'])) {
                $table->addIndex('order_item_remark_uuid', [
                    'name' => 'idx_order_item_remark_uuid',
                    'unique' => false
                ]);

                $table->update();
            }
        }
    }
}

