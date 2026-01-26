<?php

use think\migration\Migrator;

/**
 * 为外卖订单表添加 platform_total 字段
 * 
 * platform_total 字段表示平台结算总额，计算公式为：
 * platform_total = subtotal + merchant_charge_fee - merchant_discount
 * 
 * 该字段用于 Grab 等平台的准确结算金额计算
 */
class AddPlatformTotalToTakeoutOrder extends Migrator
{
    /**
     * Change Method.
     *
     * 为外卖订单表添加 platform_total 字段并更新历史数据
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('takeout_order')) {
            return;
        }

        // 获取表对象
        $table = $this->table('takeout_order');
        
        // 检查字段是否已存在
        if (!$table->hasColumn('platform_total')) {
            // 添加 platform_total 字段
            // 添加到 merchant_charge_fee 字段之后
            $table->addColumn('platform_total', 'decimal', [
                'precision' => 20,
                'scale' => 4,
                'null' => false,
                'default' => '0.0000',
                'comment' => '平台结算总额 (subtotal + merchant_charge_fee - merchant_discount)',
                'after' => 'merchant_charge_fee'
            ]);
            
            $table->update();
            
            // 更新历史订单数据
            // 使用原生 SQL 更新，计算公式：platform_total = subtotal + merchant_charge_fee - merchant_discount
            $this->execute("
                UPDATE ttpos_takeout_order 
                SET platform_total = subtotal + merchant_charge_fee - merchant_discount
                WHERE 1 = 1
            ");
        }
    }
}
