<?php
/**
 * 为扫码订单表添加订单类型字段，并更新 desk_no 字段注释
 *
 * 功能: 区分桌台扫码订单和会员端堂食订单
 */

use think\migration\Migrator;

class AddOrderTypeToH5Order extends Migrator
{
    /**
     * Change Method.
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('h5_order')) {
            $table = $this->table('h5_order');

            // 检查 order_type 字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('order_type')) {
                $table->addColumn('order_type', 'integer', [
                    'limit' => 10,
                    'default' => 0,
                    'null' => false,
                    'comment' => '订单类型, 0-桌台扫码订单 1-会员端堂食订单',
                    'after' => 'status'
                ]);
            }

            // 更新 desk_no 字段注释，增加会员端堂食订单取餐号的说明
            if ($table->hasColumn('desk_no')) {
                $table->changeColumn('desk_no', 'string', [
                    'limit' => 255,
                    'default' => '',
                    'null' => false,
                    'comment' => '桌台编号（桌台扫码订单）或取餐号（会员端堂食订单）'
                ]);
            }

            $table->update();
        }
    }
}
