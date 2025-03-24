<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddFrozenBalanceToMember extends Migrator
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
        if (!$table->hasColumn('frozen_balance')) {
            $table->addColumn('frozen_balance', 'decimal', [
                'default' => 0,
                'comment' => '冻结余额',
                'null' => false,
                'after' => 'balance',
                'precision' => 12,
                'scale' => 2,
            ]);
        }
        if (!$table->hasColumn('frozen_gift_balance')) {
            $table->addColumn('frozen_gift_balance', 'decimal', [
                'default' => 0,
                'comment' => '冻结赠送余额',
                'null' => false,
                'after' => 'gift_balance',
                'precision' => 12,
                'scale' => 2,
            ]);
        }
        $table->update();

        $table = $this->table('member_balance_log');
        if (!$table->hasColumn('processed')) {
            $table->addColumn('processed', 'integer', [
                'default' => 0,
                'comment' => '是否已处理,0-未处理 1-已处理. 用于处理会员余额变动，修改会员的余额并清0冻结的余额',
                'null' => false,
                'after' => 'gift_money',
            ]);
        }
        $table->update();
    }
}
