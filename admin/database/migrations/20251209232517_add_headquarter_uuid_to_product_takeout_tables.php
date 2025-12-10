<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddHeadquarterUuidToProductTakeoutTables extends Migrator
{
    /**
     * 为外卖相关表添加 headquarter_uuid 字段
     * - ttpos_product_package_takeout: 添加 headquarter_uuid
     * - ttpos_product_bom_takeout: 添加 headquarter_uuid
     */
    public function change()
    {
        // 1. 为 ttpos_product_package_takeout 表添加 headquarter_uuid 字段
        $table = $this->table('product_package_takeout');
        if (!$table->hasColumn('headquarter_uuid')) {
            $table->addColumn('headquarter_uuid', 'biginteger', [
                'signed' => false,
                'default' => 0,
                'comment' => '总部UUID,0表示不是总部商品',
                'after' => 'multi_language_name_uuid'
            ])
            ->update();
        }

        // 2. 为 ttpos_product_bom_takeout 表添加 headquarter_uuid 字段
        $table = $this->table('product_bom_takeout');
        if (!$table->hasColumn('headquarter_uuid')) {
            $table->addColumn('headquarter_uuid', 'biginteger', [
                'signed' => false,
                'default' => 0,
                'comment' => '总部UUID,0表示不是总部商品',
                'after' => 'product_bom_uuid'
            ])
            ->update();
        }
    }

}

