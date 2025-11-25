<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddIsRequiredAndIsDefaultToProductPackageGroupItemTable extends Migrator
{
    /**
     * 添加必选和默认选中字段到 ttpos_product_package_group_item 表
     */
    public function change()
    {
        $table = $this->table('product_package_group_item');

        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('is_required')) {
            $table->addColumn('is_required', 'integer', [
                'limit' => 10,
                'null' => false,
                'default' => 0,
                'comment' => '必选 0-不必选 1-必选',
                'after' => 'add_price'
            ])->update();
        }

        if (!$table->hasColumn('is_default')) {
            $table->addColumn('is_default', 'integer', [
                'limit' => 10,
                'null' => false,
                'default' => 0,
                'comment' => '默认选中 0-默认不选中 1-默认选中',
                'after' => 'is_required'
            ])->update();
        }
    }
}

