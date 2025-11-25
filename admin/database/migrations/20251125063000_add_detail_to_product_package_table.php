<?php

use think\migration\Migrator;
use Phinx\Db\Adapter\MysqlAdapter;

class AddDetailToProductPackageTable extends Migrator
{
    /**
     * 为商品包表新增 detail 字段（LONGTEXT）
     */
    public function change()
    {
        $table = $this->table('product_package');

        if (!$table->hasColumn('detail')) {
            $table->addColumn('detail', 'text', [
                'limit' => MysqlAdapter::TEXT_LONG,
                'comment' => '商品详情（富文本）',
                'after' => 'describe',
            ])->update();
        }
    }
}


