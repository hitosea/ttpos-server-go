<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddIsOpenMemberDiscountToSaleOrderProduct extends Migrator
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
        $table = $this->table('sale_order_product');
        if (!$table->hasColumn('is_open_member_discount')) {
            $table->addColumn('is_open_member_discount', 'tinyinteger', ['limit' => 1, 'null' => false, 'default' => 0, 'comment' => '是否开启会员折扣, 0-否 1-是', 'after' => 'is_accept_order'])
                ->update();
        }
    }
}
