<?php

use think\migration\Migrator;

/**
 * 为 member_recharge_order 和 takeout_order 添加 erp_sync_status 字段
 * 用于记录 ERP 同步状态（0=未同步 1=已入队 3=成功 4=失败）
 * 解决充值/外卖订单 SI 同步失败时无法追踪状态的问题
 */
class AddErpSyncStatusToRechargeAndTakeoutOrder extends Migrator
{
    public function up()
    {
        // 充值订单
        if ($this->hasTable('member_recharge_order')) {
            $table = $this->table('member_recharge_order');
            if (!$table->hasColumn('erp_sync_status')) {
                $table->addColumn('erp_sync_status', 'integer', [
                    'limit' => 1,
                    'default' => 0,
                    'null' => false,
                    'comment' => 'ERP同步状态: 0=未同步 1=已入队 3=成功 4=失败',
                    'after' => 'erp_payment_entry_names',
                ])->update();
            }
        }

        // 外卖订单
        if ($this->hasTable('takeout_order')) {
            $table = $this->table('takeout_order');
            if (!$table->hasColumn('erp_sync_status')) {
                $table->addColumn('erp_sync_status', 'integer', [
                    'limit' => 1,
                    'default' => 0,
                    'null' => false,
                    'comment' => 'ERP同步状态: 0=未同步 1=已入队 3=成功 4=失败',
                    'after' => 'erp_pos_invoice_resp',
                ])->update();
            }
        }
    }

    public function down()
    {
        if ($this->hasTable('member_recharge_order')) {
            $table = $this->table('member_recharge_order');
            if ($table->hasColumn('erp_sync_status')) {
                $table->removeColumn('erp_sync_status')->update();
            }
        }

        if ($this->hasTable('takeout_order')) {
            $table = $this->table('takeout_order');
            if ($table->hasColumn('erp_sync_status')) {
                $table->removeColumn('erp_sync_status')->update();
            }
        }
    }
}
