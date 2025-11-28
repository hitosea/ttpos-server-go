<?php

use think\migration\Migrator;

class AlterProductPackageDescribeToVarchar1000 extends Migrator
{
    /**
     * 将 product_package 表的 describe 字段从 VARCHAR(255) 改为 VARCHAR(1000)
     * 修复 Bug: bug-251128-001 - 编写卖点内容过长，保存报错
     */
    public function change()
    {
        $table = $this->table('product_package');

        if ($table->hasColumn('describe')) {
            $table->changeColumn('describe', 'string', [
                'limit'   => 1000,
                'null'    => false,
                'default' => '',
                'comment' => '卖点描述',
            ]);
            $table->update();
        }
    }
}

