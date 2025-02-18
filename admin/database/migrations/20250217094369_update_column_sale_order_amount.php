<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateColumnSaleOrderAmount extends Migrator
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
        // 销售订单表
        $table = $this->table('sale_order');
        if ($table->hasColumn('total_price')) {
            $table->renameColumn('total_price', 'amount')
                  ->changeColumn('amount', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '应收金额。商品未含税时，总金额=商品金额+服务费+税费。商品已含税时，总金额=商品金额（含商品消费税）+服务费+税费（只有服务费税）', 'after' => 'uuid'])
                  ->update();
        }
        
        if (!$table->hasColumn('zero_rule')) {
            $table->addColumn('zero_rule', 'tinyinteger', ['default' => 0, 'comment' => '优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数', 'after' => 'uuid']);
        }
        if (!$table->hasColumn('zero_fee')) {
            $table->addColumn('zero_fee', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '优惠折扣抹零金额。', 'after' => 'zero_rule']);
        }
        if (!$table->hasColumn('zero_checkout_rule')) {
            $table->addColumn('zero_checkout_rule', 'tinyinteger', ['default' => 0, 'comment' => '结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元', 'after' => 'zero_fee']);
        }
        if (!$table->hasColumn('zero_checkout_fee')) {
            $table->addColumn('zero_checkout_fee', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '结账抹零金额。', 'after' => 'zero_checkout_rule']);
        }
        $table->update();
    }
}
