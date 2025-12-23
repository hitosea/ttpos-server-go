<?php

use think\migration\Migrator;

class CreateTakeoutOrderPromoTable extends Migrator
{
    /**
     * 创建外卖订单促销表
     * - takeout_order_promo: 外卖订单促销信息表，支持多平台（Grab、Foodpanda 等）
     */
    public function change()
    {
        // 检查表是否已存在
        if (!$this->hasTable('takeout_order_promo')) {
            $table = $this->table('takeout_order_promo', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '外卖订单促销表(多平台)',
                'id' => false,
                'primary_key' => ['id']
            ]);

            $table
                // 基础字段
                ->addColumn('id', 'biginteger', ['signed' => false, 'identity' => true, 'comment' => '主键ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一标识'])
                ->addColumn('takeout_order_uuid', 'string', ['limit' => 255, 'default' => '', 'comment' => '外卖订单UUID'])
                
                // 平台信息
                ->addColumn('platform', 'string', ['limit' => 20, 'default' => '', 'comment' => '外卖平台: grab,lineman,etc'])
                
                // 促销信息
                ->addColumn('promo_code', 'string', ['limit' => 100, 'default' => '', 'comment' => '促销代码 (Grab: code)'])
                ->addColumn('promo_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '促销名称 (Grab: name)'])
                ->addColumn('promo_description', 'string', ['limit' => 500, 'default' => '', 'comment' => '促销描述 (Grab: description)'])
                
                // 金额信息（单位：分）
                ->addColumn('promo_amount', 'biginteger', ['signed' => true, 'default' => 0, 'comment' => '促销金额 (Grab: promoAmountInMin)'])
                ->addColumn('mex_funded_ratio', 'integer', ['limit' => 4, 'signed' => false, 'default' => 0, 'comment' => '商户承担比例 (Grab: mexFundedRatio) 百分比'])
                ->addColumn('mex_funded_amount', 'biginteger', ['signed' => true, 'default' => 0, 'comment' => '商户承担金额 (Grab: mexFundedAmount)'])
                ->addColumn('targeted_price', 'biginteger', ['signed' => true, 'default' => 0, 'comment' => '目标价格-订单小计 (Grab: targetedPrice)'])
                ->addColumn('promo_amount_in_min', 'biginteger', ['signed' => true, 'default' => 0, 'comment' => '促销金额最小单位 (Grab: promoAmountInMin)'])
                
                // 平台特定数据（JSON 格式）
                ->addColumn('platform_data', 'text', ['null' => true, 'comment' => '平台特定字段(JSON)'])
                
                // 标准字段
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])
                
                // 索引
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex(['takeout_order_uuid', 'delete_time'], ['name' => 'idx_takeout_order_uuid'])
                ->addIndex(['platform', 'promo_code', 'delete_time'], ['name' => 'idx_platform_promo'])
                
                ->create();
        }
    }
}

