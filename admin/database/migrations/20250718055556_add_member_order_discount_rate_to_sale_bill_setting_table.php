<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddMemberOrderDiscountRateToSaleBillSettingTable extends Migrator
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
        if (!$table->hasColumn('member_order_discount_rate')) {
            $table->addColumn('member_order_discount_rate', 'decimal', ['limit' => 12, 'precision' => 4, 'null' => false, 'default' => 1, 'comment' => '会员端商品价格上浮比例1%-300%', 'after' => 'member_card_discount_rate'])
                ->update();
        }
    }
}
