<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddDescribeFieldsToProductPackageTakeout extends Migrator
{
    /**
     * 为外卖商品表添加卖点字段
     */
    public function change()
    {
        $table = $this->table('product_package_takeout');
        
        // 检查字段是否已存在
        if ($table->hasColumn('describe')) {
            return;
        }

        $table->addColumn('describe', 'string', ['limit' => 255, 'default' => '', 'after' => 'image_file_uuid', 'comment' => '卖点描述'])
            ->addColumn('describe_multi_language_name_uuid', 'biginteger', ['default' => 0, 'after' => 'describe', 'comment' => '商品卖点多语言UUID'])
            ->addIndex(['describe_multi_language_name_uuid'], ['name' => 'idx_describe_multi_language_name_uuid'])
            ->update();
    }
}

