<?php

use think\migration\Migrator;

class UpdateAmountFromMemberSaleOrderTable extends Migrator
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
        $table = $this->table('member_sale_order');
        if ($table->hasColumn('amount')) {
            $table->changeColumn('amount', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00,  'comment' => '订单金额', 'after' => 'member_discount_fee']);
            $table->update();
        }
        if ($table->hasColumn('member_discount_fee')) {
            $table->changeColumn('member_discount_fee', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00,  'comment' => '会员折扣', 'after' => 'origin_product_amount']);
            $table->update();
        }
        if ($table->hasColumn('product_num')) {
            $table->changeColumn('product_num', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00,  'comment' => '商品数量.订单中商品的总数量，商品A数量2，商品B数量1，则商品数量为3', 'after' => 'payment_method_uuid']);
            $table->update();
        }
        if ($table->hasColumn('product_amount')) {
            $table->changeColumn('product_amount', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00,  'comment' => '商品金额', 'after' => 'product_num']);
            $table->update();
        }
    }
} 