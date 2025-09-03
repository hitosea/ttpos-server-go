<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddSortToProductPackageAttribute extends Migrator
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
        // 商品包属性组
        $table = $this->table('product_package_attribute_group');
        if (!$table->hasColumn('sort')) {
            $table->addColumn('sort', 'integer', ['null' => false, 'default' => 0, 'comment' => '排序', 'after' => 'product_attribute_group_uuid']);
        }
        $table->update();
        // 商品包属性值
        $table = $this->table('product_package_attribute');
        if (!$table->hasColumn('sort')) {
            $table->addColumn('sort', 'integer', ['null' => false, 'default' => 0, 'comment' => '排序', 'after' => 'is_default_selected']);
        }
        $table->update();
    }
}
