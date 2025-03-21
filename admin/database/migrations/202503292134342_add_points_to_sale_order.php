<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddPointsToSaleOrder extends Migrator
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
        if (!$table->hasColumn('gift_points')) {
            $table->addColumn('gift_points', 'decimal', [
                'precision' => 12,
                'scale' => 2,
                'default' => 0,
                'comment' => '赠送积分. 赠送积分=应收金额amount*积分赠送比例.',
            ]);
        }
        if (!$table->hasColumn('gift_points_rate')) {
            $table->addColumn('gift_points_rate', 'decimal', [
                'precision' => 12,
                'scale' => 4,
                'default' => 0,
                'comment' => '赠送积分比例. 取值范围0-1。结账后记录，不受后台改变',
            ]);
        }
        $table->update();
        $table = $this->table('member');
        if (!$table->hasColumn('frozen_point')) {
            $table->addColumn('frozen_point', 'decimal', [
                'precision' => 12,
                'scale' => 2,
                'default' => 0,
                'comment' => '冻结积分。冻结积分不能使用，在前端显示为已扣除或已增加。冻结积分可为负数。积分余额=积分+冻结积分',
            ]);
        }
        $table->update();

        $table = $this->table('member_point_log');
        if (!$table->hasColumn('related_uuid')) {
            $table->addColumn('related_uuid', 'biginteger', [
                'default' => 0,
                'comment' => '关联uuid. 表示积分变动记录关联的业务订单ID,可能是销售订单、充值订单、退款单、退货单退款金额',
            ]);
        }
        if (!$table->hasColumn('processed')) {
            $table->addColumn('processed', 'integer', [
                'default' => 0,
                'comment' => '是否已处理,0-未处理 1-已处理. 用于处理积分变动，修改会员的积分并清0冻结的积分',
            ]);
        }
        $table->update();
    }
}
