<?php
/**
 * 为采购申请物品表添加 store_snapshot_quantity 字段
 * 记录提交采购申请时的门店库存快照（默认销售单位）
 */

use think\migration\Migrator;

class AddStoreSnapshotQuantityToPurchaseOrderItem extends Migrator
{
    public function change()
    {
        if ($this->hasTable('purchase_order_item')) {
            $table = $this->table('purchase_order_item');
            if (!$table->hasColumn('store_snapshot_quantity')) {
                $table->addColumn('store_snapshot_quantity', 'decimal', [
                    'precision' => 14,
                    'scale' => 4,
                    'default' => 0,
                    'null' => false,
                    'comment' => '申请时门店库存（默认销售单位，如不存在，取基准单位）',
                    'after' => 'supplier_erp_code',
                ])->update();
            }
        }
    }
}
