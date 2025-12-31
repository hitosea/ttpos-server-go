<?php

use think\migration\Migrator;

class RenameTakeoutOrderItemFields extends Migrator
{
    /**
     * 重命名外卖订单商品表字段
     * - ttpos_takeout_order_item.ttpos_product_uuid -> ttpos_product_package_uuid
     * - ttpos_takeout_order_item_modifier.ttpos_flavor_uuid -> ttpos_flavor_product_bom_uuid
     * - ttpos_takeout_order_item_modifier.ttpos_product_uuid -> ttpos_product_package_uuid
     */
    public function change()
    {
        // 1. 修改 ttpos_takeout_order_item 表
        if ($this->hasTable('takeout_order_item')) {
            $table = $this->table('takeout_order_item');
            // 检查旧字段是否存在，新字段是否不存在
            if ($table->hasColumn('ttpos_product_uuid') && !$table->hasColumn('ttpos_product_package_uuid')) {
                $table->renameColumn('ttpos_product_uuid', 'ttpos_product_package_uuid')->save();
            }
        }

        // 2. 修改 ttpos_takeout_order_item_modifier 表
        if ($this->hasTable('takeout_order_item_modifier')) {
            $table = $this->table('takeout_order_item_modifier');
            
            // 重命名 ttpos_flavor_uuid -> ttpos_flavor_product_bom_uuid
            if ($table->hasColumn('ttpos_flavor_uuid') && !$table->hasColumn('ttpos_flavor_product_bom_uuid')) {
                $table->renameColumn('ttpos_flavor_uuid', 'ttpos_flavor_product_bom_uuid')->save();
            }
            
            // 重命名 ttpos_product_uuid -> ttpos_product_package_uuid
            if ($table->hasColumn('ttpos_product_uuid') && !$table->hasColumn('ttpos_product_package_uuid')) {
                $table->renameColumn('ttpos_product_uuid', 'ttpos_product_package_uuid')->save();
            }
        }
    }
}

