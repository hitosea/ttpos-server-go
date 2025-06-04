<?php

use think\migration\Migrator;

class AddColumnSaleOrder extends Migrator
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
        if (!$table->hasColumn('pay_points')) {
            $table->addColumn('pay_points', 'decimal', ['default' => 0, 'comment' => '抵扣积分,用了多少积分进行抵扣', 'after' => 'zero_checkout_rule', 'precision' => 12, 'scale' => 2]);
        }
        if (!$table->hasColumn('pay_points_amount')) {
            $table->addColumn('pay_points_amount', 'decimal', ['default' => 0, 'comment' => '抵扣金额,关联销售订单的积分抵扣了多少金额', 'after' => 'pay_points', 'precision' => 12, 'scale' => 2]);
        }
        if (!$table->hasColumn('points_exchange_rate')) {
            $table->addColumn('points_exchange_rate', 'decimal', ['default' => 0, 'comment' => '积分抵扣汇率,1积分抵扣多少元', 'after' => 'pay_points_amount', 'precision' => 12, 'scale' => 4]);
        }
        // 判断是否存在字段 'unit'，如果不存在则添加
        if (!$table->hasColumn('unit')) {
            $table->addColumn('unit', 'string', ['default' => '', 'comment' => '积分抵扣金额的单位,$-美元 ￥-人民币,用于显示订单当时积分抵扣的金额价值', 'after' => 'member_balance', 'limit' => 255]);
        }
        $table->update();

        $table = $this->table('sale_bill_setting');
        if (!$table->hasColumn('open_points_exchange')) {
            $table->addColumn('open_points_exchange', 'integer', ['default' => 0, 'comment' => '是否开启积分抵扣, 0-不开启 1-开启', 'after' => 'discount_type']);
        }
        if (!$table->hasColumn('points_exchange_rate')) {
            $table->addColumn('points_exchange_rate', 'decimal', ['default' => 0, 'comment' => '积分抵扣汇率,1积分抵扣多少元', 'after' => 'open_points_exchange', 'precision' => 12, 'scale' => 4]);
        }
        if (!$table->hasColumn('auto_points_exchange')) {
            $table->addColumn('auto_points_exchange', 'integer', ['default' => 0, 'comment' => '积分抵扣类型,0-手动抵扣 1-自动抵扣', 'after' => 'points_exchange_rate']);
        }
        $table->update();

    }
}
