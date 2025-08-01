<?php

use think\migration\Migrator;
use think\migration\db\Column;
use think\facade\Db;

class AddMemberFieldsToShopMaster extends Migrator
{
    // 迁移目标
    const TARGET = 'shop_master';

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
            // 修改会员表，添加游客相关字段
            $table = $this->table('member');
                
            // 添加是否游客字段
            if (!$table->hasColumn('is_visitor')) {
                $table->addColumn('is_visitor', 'integer', [
                    'default' => 0, 
                    'null' => false, 
                    'comment' => '是否游客,0-否 1-是',
                    'after' => 'phone'
                ]);
                $table->update();
            }
            
            // 添加设备ID字段
            if (!$table->hasColumn('device_id')) {
                $table->addColumn('device_id', 'string', [
                    'limit' => 255, 
                    'default' => '', 
                    'null' => false, 
                    'comment' => '设备ID,用于标识游客',
                    'after' => 'is_visitor'
                ]);
                
                // 添加设备ID索引
                $table->addIndex(['device_id'], [
                    'name' => 'idx_device_id',
                    'unique' => false
                ]);
                $table->update();
            }

            // 添加累计获取积分字段
            if (!$table->hasColumn('accumulated_get_point')) {
                $table->addColumn('accumulated_get_point', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '累计获取积分', 'after' => 'frozen_point']);
                $table->update();

                // 初始化数据
                $db = Db::connect(Db::getConfig('default'), true);
                $db->name('member')->where('accumulated_get_point', 0)->update(['accumulated_get_point' => Db::raw('point + frozen_point')]);
            }

            // 添加累计消费获取积分字段
            if (!$table->hasColumn('accumulated_consumption_get_point')) {
                $table->addColumn('accumulated_consumption_get_point', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '累计消费获取积分(只存消费赠送积分，不存充值与活动赠送积分)', 'after' => 'accumulated_get_point']);
                $table->update();

                // 初始化数据
                $db = Db::connect(Db::getConfig('default'), true);
                $db->name('member')->where('accumulated_consumption_get_point', 0)->update(['accumulated_consumption_get_point' => Db::raw('point + frozen_point')]);
            }
        } catch (\Exception $e) {
            
        }

        try {
            // 会员积分日志变更字段
            $table = $this->table('member_point_log');
            if ($table->hasColumn('scene')) {
                $table->changeColumn('scene', 'integer', ['limit' => 11, 'null' => false, 'default' => 0, 'comment' => '场景,10-用户充值 20-订单赠送 30-管理员操作 40-退款扣除 60-订单反结账 70-充值赠送 80-充值反结账 90-扣减 100-收银机、点餐助手发卡赠送 110-积分抵扣 120-积分抵扣反结账 130-营销活动赠送', 'after' => 'member_uuid']);
                $table->update();
            }

            // 会员积分日志添加备注字段
            if (!$table->hasColumn('remark')) {
                $table->addColumn('remark', 'text', ['comment' => '备注', 'after' => 'related_uuid'])
                    ->update();
            }
        } catch (\Exception $e) {

        }

        try {
            // 会员余额日志添加备注字段
            $table = $this->table('member_balance_log');
            if (!$table->hasColumn('remark')) {
                $table->addColumn('remark', 'text', ['comment' => '备注', 'after' => 'related_uuid'])
                    ->update();
            }     
        } catch (\Exception $e) {

        }
    }
}
