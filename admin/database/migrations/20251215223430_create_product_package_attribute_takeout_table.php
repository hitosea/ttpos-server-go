<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateProductPackageAttributeTakeoutTable extends Migrator
{
    /**
     * 创建外卖属性价格表
     */
    public function change()
    {
        // 检查表是否已存在
        if ($this->hasTable('product_package_attribute_takeout')) {
            return;
        }
        
        $table = $this->table('product_package_attribute_takeout', ['engine' => 'InnoDB', 'collation' => 'utf8mb4_unicode_ci', 'comment' => '外卖属性价格表']);
        $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一标识'])
            ->addColumn('product_package_takeout_uuid', 'biginteger', ['default' => 0, 'comment' => '外卖商品UUID，关联 ttpos_product_package_takeout.uuid'])
            ->addColumn('product_package_attribute_uuid', 'biginteger', ['default' => 0, 'comment' => '店内商品属性 UUID，关联 ttpos_product_package_attribute.uuid'])
            ->addColumn('headquarter_uuid', 'biginteger', ['default' => 0, 'comment' => '总部UUID,0表示不是总部商品'])
            ->addColumn('price', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => '0.0000', 'comment' => '外卖属性价格'])
            ->addColumn('delete_time', 'biginteger', ['default' => 0, 'comment' => '删除时间，0表示未删除'])
            ->addColumn('create_time', 'biginteger', ['default' => 0, 'comment' => '创建时间'])
            ->addColumn('update_time', 'biginteger', ['default' => 0, 'comment' => '更新时间'])
            ->addIndex(['uuid'], ['name' => 'idx_uuid'])
            ->addIndex(['product_package_takeout_uuid'], ['name' => 'idx_product_package_takeout_uuid'])
            ->addIndex(['product_package_attribute_uuid'], ['name' => 'idx_product_package_attribute_uuid'])
            ->addIndex(['headquarter_uuid'], ['name' => 'idx_headquarter_uuid'])
            ->addIndex(['delete_time'], ['name' => 'idx_delete_time'])
            ->create();
    }
}

