<?php

use think\migration\Migrator;

/**
 * 为会员充值订单表添加反结账次数字段
 *
 * 在 ttpos_member_recharge_order 表中添加：
 * - reverse_settle_count: 反结账次数
 */
class AddReverseSettleCountToMemberRechargeOrder extends Migrator
{
    /**
     * Change Method.
     *
     * 为会员充值订单表添加反结账次数字段
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('member_recharge_order')) {
            return;
        }

        // 获取表对象
        $table = $this->table('member_recharge_order');

        // 添加 reverse_settle_count 字段 - 反结账次数
        if (!$table->hasColumn('reverse_settle_count')) {
            $table->addColumn('reverse_settle_count', 'integer', [
                'limit' => 11,
                'null' => false,
                'default' => 0,
                'comment' => '反结账次数',
                'after' => 'erp_products_invoice_name'
            ]);
        }

        $table->update();
    }
}
