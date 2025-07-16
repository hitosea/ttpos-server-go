<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddMemberSaleOrderTable extends Migrator
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
        $table = $this->table('member_sale_order');
        if (!$this->hasTable('member_sale_order')) {
            $table = $this->table('member_sale_order', ['comment' => '会员销售订单表']);
            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '会员销售订单ID'])
            ->addColumn('status', 'integer', ['null' => false, 'default' => 1, 'comment' => '订单状态 0-选购中 1-待支付 2-待商家接单 3-商家备餐中 4-待骑手接单 5-骑手正在赶往商家 6-骑手配送中 7-已完成 8-已取消'])
            ->addColumn('delivery_distance', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '配送距离，单位km'])
            ->addColumn('remark', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '订单备注'])
            ->addColumn('delivery_fee_amount', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '配送费'])
            ->addColumn('delivery_fee_distance', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '配送距离'])
            ->addColumn('delivery_fee_min_fee', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '起步配送费'])
            ->addColumn('delivery_fee_base_fee', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '基础配送费'])
            ->addColumn('delivery_fee_per_km', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '每公里配送费'])
            ->addColumn('related_order_no', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '关联订单号,skootar、grab等第三方平台上的订单号'])
            ->addColumn('related_order_type', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '关联订单类型,skootar、grab'])
            ->addColumn('rider_name', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '骑手名称'])
            ->addColumn('rider_phone', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '骑手电话'])
            ->addColumn('rider_latitude', 'decimal', ['precision' => 12, 'scale' => 6, 'null' => false, 'default' => 0.00, 'comment' => '骑手纬度'])
            ->addColumn('rider_longitude', 'decimal', ['precision' => 12, 'scale' => 6, 'null' => false, 'default' => 0.00, 'comment' => '骑手经度'])
            ->addColumn('remaining_distance', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '剩余距离'])
            ->addColumn('pay_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '支付完成时间（时间戳）'])
            ->addColumn('accept_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '商家接单时间（时间戳）'])
            ->addColumn('cook_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '商家备餐完成时间（时间戳）'])
            ->addColumn('rider_accept_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '骑手接单时间（时间戳）'])
            ->addColumn('rider_start_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '骑手开始配送时间（时间戳）'])
            ->addColumn('finish_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '骑手送达时间（时间戳）'])
            ->addColumn('expected_finish_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '预计送达时间（时间戳）'])
            ->addColumn('cancel_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '取消时间（时间戳）'])
            ->addColumn('create_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '创建时间(时间戳)，前端提交订单的时间'])
            ->addColumn('update_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '更新时间(时间戳)'])
            ->addColumn('delete_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '删除时间(时间戳)'])
            ->create();
        }
    }
}
