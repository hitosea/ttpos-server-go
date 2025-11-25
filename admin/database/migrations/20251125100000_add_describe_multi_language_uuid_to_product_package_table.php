<?php

use think\migration\Migrator;

class AddDescribeMultiLanguageUuidToProductPackageTable extends Migrator
{
    /**
     * 为商品包表新增商品卖点多语言字段
     */
    public function change()
    {
        $table = $this->table('product_package');

        if (!$table->hasColumn('describe_multi_language_name_uuid')) {
            $table->addColumn('describe_multi_language_name_uuid', 'biginteger', [
                'default' => 0,
                'comment' => '商品卖点多语言UUID',
                'after'   => 'describe',
                'signed'  => false,
            ])->addIndex(['describe_multi_language_name_uuid'], [
                'name' => 'idx_describe_multi_lang_uuid',
            ])->update();
        }
    }
}

