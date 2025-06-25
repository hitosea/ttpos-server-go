<?php

use think\migration\Migrator;

class AddColumeSaleOrderCouponAmount extends Migrator
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
        $table = $this->table('sale_order');
        if (!$table->hasColumn('coupon_amount')) {
            $table->addColumn('coupon_amount', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '优惠券抵扣金额,抵扣了多少金额', 'after' => 'auto_points_exchange']);
            $table->update();
        }
    }
}

