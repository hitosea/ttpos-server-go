<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddSortToProductPackageGroupTable extends Migrator
{
    /**
     * 添加排序字段到 ttpos_product_package_group 表
     */
    public function change()
    {
        $table = $this->table('product_package_group');

        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('sort')) {
            $table->addColumn('sort', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '排序字段，数值越小越靠前',
                'after' => 'optional_count'
            ])->update();
        }
    }
}

