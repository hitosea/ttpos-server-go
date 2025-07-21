<?php

use think\migration\Migrator;
use think\migration\db\Column;

class addOpenOverallDiscountToSaleOrderBuffetCustomerType extends Migrator
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
        $table = $this->table('sale_order_buffet_customer_type');
        if (!$table->hasColumn('open_overall_discount')) {
            $table->addColumn('open_overall_discount', 'biginteger', ['signed' => false, 'null' => false, 'default' => 1, 'comment' => '是否开启整单折扣, 0-否 1-是', 'after' => 'custom_discount_rate']);
            $table->update();
        }
    }
}
