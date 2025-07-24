<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddReverseSettleCountToSaleBill extends Migrator
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
        // 添加字段
        $table = $this->table('sale_bill');
        if (!$table->hasColumn('reverse_settle_count')) {
            $table->addColumn('reverse_settle_count', 'integer', ['limit' => 11, 'null' => false, 'default' => 0, 'comment' => '反结账次数', 'after' => 'is_kitchen_confirm']);
            $table->update();
        }

        // 添加字段
        $table = $this->table('marketing_activity_record');
        if (!$table->hasColumn('reward_value')) {
            $table->addColumn('reward_value', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '奖励值', 'after' => 'last_reward_time']);
            $table->update();
        }
    }
}
