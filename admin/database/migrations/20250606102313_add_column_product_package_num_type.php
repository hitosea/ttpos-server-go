<?php

use think\migration\Migrator;

class AddColumnProductPackageNumType extends Migrator
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
        if (!$table->hasColumn('num_type')) {
            $table->addColumn('num_type', 'integer', ['limit' => 10, 'default' => 0, 'comment' => '数量计算方法, 0-整数 1-小数', 'after' => 'deduct_stock_type']);
        }
        $table->update();
        $table = $this->table('sale_order_product');
        if ($table->hasColumn('num')) {
            $table->changeColumn('num', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '商品数量。不能减为0，当数量为1再减时，标记删除', 'after' => 'multi_language_name_uuid']);
        }
        $table->update();
    }
}

