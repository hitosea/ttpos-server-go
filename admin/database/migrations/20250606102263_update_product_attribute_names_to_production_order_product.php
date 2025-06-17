<?php

use think\migration\Migrator;

class UpdateProductAttributeNamesToProductionOrderProduct extends Migrator
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
        $table = $this->table('production_order_product');
        if ($table->hasColumn('product_attribute_names')) {
            $table->changeColumn('product_attribute_names', 'text', ['default' => '', 'comment' => '商品属性名称,多个属性名用逗号分隔,不随后台改变', 'after' => 'flavor_name']);
            $table->update();
        }
    }
}
