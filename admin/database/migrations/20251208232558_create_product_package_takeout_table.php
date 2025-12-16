<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateProductPackageTakeoutTable extends Migrator
{
    /**
     * 创建外卖商品相关表
     * - ttpos_product_package_takeout: 外卖商品表
     * - ttpos_product_bom_takeout: 外卖规格价格表
     */
    public function change()
    {
        // 1. 创建外卖商品表 ttpos_product_package_takeout
        if (!$this->hasTable('product_package_takeout')) {
            $table = $this->table('product_package_takeout', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '外卖商品表，存储商品的外卖专属信息',
                'id' => false,
                'primary_key' => ['id']
            ]);

            $table
                // 基础字段
                ->addColumn('id', 'biginteger', ['signed' => false, 'identity' => true, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'UUID'])

                // 关联字段
                ->addColumn('product_package_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '商品包UUID，关联 ttpos_product_package.uuid'])
                ->addColumn('multi_language_name_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '多语言名称ID'])
                ->addColumn('headquarter_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '总部UUID,0表示不是总部商品'])

                // 商品信息字段
                ->addColumn('name', 'text', ['null' => true, 'comment' => '商品包名称'])
                ->addColumn('product_type', 'integer', ['limit' => 4, 'signed' => false, 'default' => 0, 'comment' => '商品类型, 0-商品 1-套餐'])

                // 外卖专属字段
                ->addColumn('takeout_type', 'integer', ['limit' => 4, 'signed' => false, 'default' => 1, 'comment' => '外卖类型 1-Grab 2-FoodPanda 3-其他'])
                ->addColumn('status', 'integer', ['limit' => 4, 'signed' => false, 'default' => 0, 'comment' => '外卖状态 0-下架 1-上架'])
                ->addColumn('category_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '外卖分类UUID'])
                ->addColumn('special_category_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '外卖特色分类UUID'])
                ->addColumn('image_file_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '外卖商品图片UUID'])

                // 时间字段
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])

                // 索引
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'idx_uuid'])
                ->addIndex(['takeout_type'], ['name' => 'idx_takeout_type'])
                ->addIndex(['status'], ['name' => 'idx_status'])
                ->addIndex(['delete_time'], ['name' => 'idx_delete_time'])

                ->create();
        }

        // 2. 创建外卖规格价格表 ttpos_product_bom_takeout
        if (!$this->hasTable('product_bom_takeout')) {
            $table = $this->table('product_bom_takeout', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '外卖规格价格表',
                'id' => false,
                'primary_key' => ['id']
            ]);

            $table
                // 基础字段
                ->addColumn('id', 'biginteger', ['signed' => false, 'identity' => true, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'UUID'])

                // 关联字段
                ->addColumn('product_package_takeout_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '外卖商品UUID，关联 ttpos_product_package_takeout.uuid'])
                ->addColumn('product_bom_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '店内商品BOM UUID，关联 ttpos_product_bom.uuid'])
                ->addColumn('headquarter_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '总部UUID,0表示不是总部商品'])

                // 价格字段（核心）
                ->addColumn('price', 'decimal', ['precision' => 22, 'scale' => 4, 'default' => 0.0000, 'comment' => '外卖规格价格'])

                // 时间字段
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])

                // 索引
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'idx_uuid'])
                ->addIndex(['product_package_takeout_uuid', 'product_bom_uuid'], ['unique' => true, 'name' => 'idx_takeout_bom'])
                ->addIndex(['product_package_takeout_uuid'], ['name' => 'idx_product_package_takeout_uuid'])
                ->addIndex(['product_bom_uuid'], ['name' => 'idx_product_bom_uuid'])
                ->addIndex(['delete_time'], ['name' => 'idx_delete_time'])

                ->create();
        }
    }
}
