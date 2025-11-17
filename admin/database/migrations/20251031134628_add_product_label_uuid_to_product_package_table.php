<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddProductLabelUuidToProductPackageTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     *
     * The following commands can be used in this method and Phinx will
     * automatically reverse them when rolling back:
     *
     *    createTable
     *    renameTable
     *    addColumn
     *    renameColumn
     *    addIndex
     *    addForeignKey
     *
     * Remember to call "create()" or "update()" and NOT "save()" when working
     * with the Table class.
     */
    public function change()
    {
        $table = $this->table('product_package');
        
        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('product_label_uuid')) {
            $table->addColumn('product_label_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '商品标签UUID', 'after' => 'headquarter_uuid'])
                ->update();
        }
    }
}

