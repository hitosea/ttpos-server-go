<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddGroupTypeAndOptionalCountToProductPackageGroupTable extends Migrator
{
    /**
     * 添加分组类型和可选数量字段到 ttpos_product_package_group 表
     */
    public function change()
    {
        $table = $this->table('product_package_group');

        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('group_type')) {
            $table->addColumn('group_type', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '分组类型 0-固定 1-可选',
                'after' => 'product_package_uuid'
            ])->update();
        }

        if (!$table->hasColumn('optional_count')) {
            $table->addColumn('optional_count', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '可选数量，表示本组商品中要求选择多少个商品',
                'after' => 'group_type'
            ])->update();
        }
    }
}

