<?php

use think\facade\Db;
use think\migration\Migrator;

class AddAccumulatedGetPointToMember extends Migrator
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
        try {
            // 添加字段
            if ($this->hasTable('member')) {
                $table = $this->table('member');
                if (!$table->hasColumn('accumulated_get_point')) {
                    $table->addColumn('accumulated_get_point', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '累计获取积分', 'after' => 'frozen_point']);
                    $table->update();

                    // 初始化数据
                    $db = Db::connect(Db::getConfig('default'), true);
                    $db->name('member')->where('accumulated_get_point', 0)->update(['accumulated_get_point' => Db::raw('point + frozen_point')]);
                }

                // 添加字段
                $table = $this->table('member');
                if (!$table->hasColumn('accumulated_consumption_get_point')) {
                    $table->addColumn('accumulated_consumption_get_point', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '累计消费获取积分(只存消费赠送积分，不存充值与活动赠送积分)', 'after' => 'accumulated_get_point']);
                    $table->update();

                    // 初始化数据
                    $db = Db::connect(Db::getConfig('default'), true);
                    $db->name('member')->where('accumulated_consumption_get_point', 0)->update(['accumulated_consumption_get_point' => Db::raw('point + frozen_point')]);
                }
            }
        } catch (\Exception $e) {

        }

        try {
            // 变更字段
            if ($this->hasTable('member_point_log')) {
                $table = $this->table('member_point_log');
                if ($table->hasColumn('scene')) {
                    $table->changeColumn('scene', 'integer', ['limit' => 11, 'null' => false, 'default' => 0, 'comment' => '场景,10-用户充值 20-订单赠送 30-管理员操作 40-退款扣除 60-订单反结账 70-充值赠送 80-充值反结账 90-扣减 100-收银机、点餐助手发卡赠送 110-积分抵扣 120-积分抵扣反结账 130-营销活动赠送', 'after' => 'member_uuid']);
                    $table->update();
                }
            }
        } catch (\Exception $e) {

        }

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
