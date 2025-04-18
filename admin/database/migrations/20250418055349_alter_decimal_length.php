<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AlterDecimalLength extends Migrator
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
        $table = $this->table('member');
        if ($table->hasColumn('point')) {
            $table->changeColumn('point', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '积分', 'after' => 'birthday'])->update();
        }
        if ($table->hasColumn('frozen_point')) {
            $table->changeColumn('frozen_point', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '冻结积分。冻结积分不能使用，在前端显示为已扣除或已增加。冻结积分可为负数。积分余额=积分+冻结积分', 'after' => 'point'])->update();
        }
        if ($table->hasColumn('accumulated_consumption_amount')) {
            $table->changeColumn('accumulated_consumption_amount', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '累计消费金额', 'after' => 'frozen_point'])->update();
        }
        if ($table->hasColumn('balance')) {
            $table->changeColumn('balance', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '余额', 'after' => 'consumption_count'])->update();
        }
        if ($table->hasColumn('frozen_balance')) {
            $table->changeColumn('frozen_balance', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '冻结余额。冻结余额不能使用，在前端显示为已扣除或已增加。冻结余额可为负数。会员余额=余额+冻结余额', 'after' => 'balance'])->update();
        }
        if ($table->hasColumn('gift_balance')) {
            $table->changeColumn('gift_balance', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '赠送账户余额', 'after' => 'frozen_balance'])->update();
        }
        if ($table->hasColumn('frozen_gift_balance')) {
            $table->changeColumn('frozen_gift_balance', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '冻结赠送账户余额。冻结赠送账户余额不能使用，在前端显示为已扣除或已增加。冻结赠送账户余额可为负数。赠送账户余额=赠送账户余额+冻结赠送账户余额', 'after' => 'gift_balance'])->update();
        }
        if ($table->hasColumn('accumulated_recharge_amount')) {
            $table->changeColumn('accumulated_recharge_amount', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '累计充值金额', 'after' => 'frozen_gift_balance'])->update();
        }

        $table = $this->table('member_balance_log');
        if ($table->hasColumn('money')) {
            $table->changeColumn('money', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '变动金额,负数:减余额 正数:加余额。包含赠送余额', 'after' => 'scene'])->update();
        }
        if ($table->hasColumn('gift_money')) {
            $table->changeColumn('gift_money', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '变动赠送金额', 'after' => 'money'])->update();
        }

        $table = $this->table('member_card_log');
        if ($table->hasColumn('price')) {
            $table->changeColumn('price', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '价格,会员卡价格,不随后台改变,记录领取时的价格', 'after' => 'uuid'])->update();
        }
        if ($table->hasColumn('give_money')) {
            $table->changeColumn('give_money', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '赠送余额', 'after' => 'member_uuid'])->update();
        }
        if ($table->hasColumn('give_point')) {
            $table->changeColumn('give_point', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '赠送积分', 'after' => 'give_money'])->update();
        }

        $table = $this->table('member_card_type');
        if ($table->hasColumn('price')) {
            $table->changeColumn('price', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '会员卡类型名称', 'after' => 'uuid'])->update();
        }
        if ($table->hasColumn('open_point_num')) {
            $table->changeColumn('open_point_num', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '开卡赠送积分数', 'after' => 'open_point'])->update();
        }
        if ($table->hasColumn('open_money_num')) {
            $table->changeColumn('open_money_num', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '开卡赠送余额数', 'after' => 'open_money'])->update();
        }
    }
}
