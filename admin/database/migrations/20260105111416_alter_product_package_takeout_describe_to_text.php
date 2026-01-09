<?php

use think\migration\Migrator;

class AlterProductPackageTakeoutDescribeToText extends Migrator
{
    /**
     * 将 product_package_takeout 表的 describe 字段从 VARCHAR(255) 改为 TEXT
     * 修复 Bug: 添加外卖商品失败，describe 字段数据过长
     */
    public function change()
    {
        if (!$this->hasTable('product_package_takeout')) {
            return;
        }
        
        $table = $this->table('product_package_takeout');
        if ($table->hasColumn('describe')) {
            $table->changeColumn('describe', 'text', ['null' => false, 'comment' => '卖点描述']);
            $table->update();
        }
    }
}

