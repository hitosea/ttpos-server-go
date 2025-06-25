<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddTableSaleOrderCoupon extends Migrator
{
    public function change()
    {
        if (!$this->hasTable('sale_order_coupon')) {
            $table = $this->table('sale_order_coupon', ['comment' => '销售订单优惠券表']);
            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '销售订单优惠券ID'])
                ->addColumn('coupon_amount', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '优惠券抵扣金额，实际抵扣金额'])
                ->addColumn('coupon_origin_amount', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '优惠券原始金额(面值)'])
                ->addColumn('coupon_requirement', 'string', ['limit' => 20, 'default' => '', 'comment' => '优惠券的类型，none-所有人可用，但一个saleBill只能用一张优惠券;marketing-会员通过营销活动获的优惠券；'])
                ->addColumn('member_coupon_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '会员的优惠券uuid,表示该订单使用会员的哪个优惠券。none时有值'])
                ->addColumn('marketing_coupon_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '营销优惠券uuid,表示该订单使用营销的哪个优惠券。marketing时有值'])
                ->addColumn('sale_order_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '销售订单ID'])
                ->addColumn('create_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '删除时间(时间戳)'])
                ->create();
        }
    }
} 