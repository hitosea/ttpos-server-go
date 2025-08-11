<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddProductPackageAttributeUuidToSaleOrderProductAttributeTable extends Migrator
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
        $table = $this->table('sale_order_product_attribute');
        
        // 检查字段是否已存在
        $hasColumn = $table->hasColumn('product_package_attribute_uuid');
        if (!$hasColumn) {
            // 添加商品包属性UUID字段
            $table->addColumn('product_package_attribute_uuid', 'biginteger', [
                'null' => false,
                'default' => 0,
                'comment' => '商品包属性ID',
                'after' => 'product_attribute_uuid'
            ])->update();
        }
    }
} 