<?php

use think\facade\Db;
use think\migration\Migrator;

class AddRewardTypeToMarketingActivity extends Migrator
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
        $table = $this->table('marketing_activity');
        if (!$table->hasColumn('reward_type')) {
            $table->addColumn('reward_type', 'integer', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '奖励类型,0:优惠券,1:积分', 'after' => 'end_time']);
            $table->addColumn('reward_value', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '奖励值', 'after' => 'reward_type']);
            $table->addColumn('is_send_sms', 'integer', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '是否发送短信通知,0:否,1:是', 'after' => 'reward_value']);
            $table->update();
        }
    }
}
