<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateMemberCouponTable extends Migrator
{
    public function change()
    {
        if (!$this->hasTable('member_coupon')) {
            $table = $this->table('member_coupon', ['comment' => '会员优惠券表']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID'])
                ->addColumn('member_uuid', 'biginteger', ['default' => 0, 'comment' => '会员uuid'])
                ->addColumn('coupon_uuid', 'biginteger', ['default' => 0, 'comment' => '优惠券uuid'])
                ->addColumn('name', 'string', ['limit' => 50, 'default' => '', 'comment' => '优惠券名称'])
                ->addColumn('deduction_type', 'string', ['limit' => 20, 'default' => '', 'comment' => '抵扣类型: taxed - 税后抵扣'])
                ->addColumn('type', 'string', ['limit' => 20, 'default' => '', 'comment' => '优惠券类型: deduction - 抵扣券'])
                ->addColumn('day_start_time', 'string', ['limit' => 5, 'default' => '', 'comment' => '每日适用时段开始时间, hh:mm 格式'])
                ->addColumn('day_end_time', 'string', ['limit' => 5, 'default' => '', 'comment' => '每日适用时段结束时间, hh:mm 格式'])
                ->addColumn('valid_start_time', 'integer', ['default' => 0, 'comment' => '优惠券有效开始时间, requirement = none 时有效'])
                ->addColumn('valid_end_time', 'integer', ['default' => 0, 'comment' => '优惠券有效结束时间, requirement = none 时有效'])
                ->addColumn('amount', 'decimal', ['precision' => 14, 'scale' => 2, 'default' => 0, 'comment' => '优惠券面值'])
                ->addColumn('status', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '优惠券状态 0未使用 1已使用'])
                ->addColumn('start_time', 'integer', ['default' => 0, 'comment' => '优惠券开始时间'])
                ->addColumn('end_time', 'integer', ['default' => 0, 'comment' => '优惠券结束时间'])
                ->addColumn('use_time', 'integer', ['default' => 0, 'comment' => '优惠券使用时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->create();
        }

        if (!$this->hasTable('member_coupon_use_record')) {
            $table = $this->table('member_coupon_use_record', ['comment' => '会员优惠券使用记录表']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID'])
                ->addColumn('member_uuid', 'biginteger', ['default' => 0, 'comment' => '会员uuid'])
                ->addColumn('coupon_uuid', 'biginteger', ['default' => 0, 'comment' => '优惠券uuid'])
                ->addColumn('use_order_uuid', 'biginteger', ['default' => 0, 'comment' => '优惠券使用订单uuid'])
                ->addColumn('use_order_amount', 'decimal', ['precision' => 14, 'scale' => 2, 'default' => 0, 'comment' => '优惠券使用订单金额'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->create();
        }

        if ($this->hasTable('marketing_coupon_record')) {
            $table = $this->table('marketing_coupon_record');
            if (!$table->hasColumn('member_uuid')) {
                $table->addColumn('member_uuid', 'biginteger', ['default' => 0, 'comment' => '会员uuid', 'after' => 'coupon_uuid']);
            }
            if (!$table->hasColumn('activity_uuid')) {
                $table->addColumn('activity_uuid', 'biginteger', ['default' => 0, 'comment' => '活动uuid', 'after' => 'coupon_uuid']);
            }
            $table->update();
        }
    }
} 