<?php

use think\migration\Migrator;

class CreateTableTakeoutOrderReceivers extends Migrator
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
        // 外卖订单收货人信息表
        if (!$this->hasTable('takeout_order_receiver')) {
            $table = $this->table('takeout_order_receiver', [
                'id' => 'id',
                'comment' => '外卖订单收货人信息表'
            ]);
            
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => 'UUID'])
                ->addColumn('takeout_order_uuid', 'biginteger', ['default' => 0, 'comment' => '外卖订单UUID'])
                ->addColumn('platform', 'string', ['limit' => 50, 'default' => '', 'comment' => '平台名称(grab/lineman/foodpanda等)'])
                ->addColumn('receiver_name', 'string', ['limit' => 100, 'default' => '', 'comment' => '收货人姓名'])
                ->addColumn('receiver_phones', 'string', ['limit' => 50, 'default' => '', 'comment' => '收货人电话'])
                ->addColumn('unit_number', 'string', ['limit' => 50, 'default' => '', 'comment' => '单元号/门牌号'])
                ->addColumn('delivery_instruction', 'string', ['limit' => 500, 'default' => '', 'comment' => '配送说明'])
                ->addColumn('poi_source', 'string', ['limit' => 50, 'default' => '', 'comment' => 'POI来源(GRAB/GOOGLE/FACEBOOK等)'])
                ->addColumn('poi_id', 'string', ['limit' => 100, 'default' => '', 'comment' => 'POI ID'])
                ->addColumn('address', 'string', ['limit' => 500, 'default' => '', 'comment' => '完整地址'])
                ->addColumn('postcode', 'string', ['limit' => 20, 'default' => '', 'comment' => '邮政编码'])
                ->addColumn('latitude', 'decimal', ['precision' => 10, 'scale' => 7, 'default' => '0.0000000', 'comment' => '纬度'])
                ->addColumn('longitude', 'decimal', ['precision' => 10, 'scale' => 7, 'default' => '0.0000000', 'comment' => '经度'])
                ->addColumn('create_time', 'biginteger', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'biginteger', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'biginteger', ['default' => 0, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'idx_uuid'])
                ->addIndex(['takeout_order_uuid'], ['unique' => true, 'name' => 'idx_takeout_order_uuid'])
                ->addIndex(['delete_time'], ['name' => 'idx_delete_time'])
                ->create();
        }
    }
}

