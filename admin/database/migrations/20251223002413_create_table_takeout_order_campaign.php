<?php

use think\migration\Migrator;

class CreateTableTakeoutOrderCampaign extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     */
    public function change()
    {
        // 外卖订单活动表
        if (!$this->hasTable('takeout_order_campaign')) {
            $table = $this->table('takeout_order_campaign', [
                'id' => 'id',
                'comment' => '外卖订单活动表'
            ]);
            
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => 'UUID'])
                ->addColumn('takeout_order_uuid', 'biginteger', ['default' => 0, 'comment' => '外卖订单UUID'])
                ->addColumn('platform', 'string', ['limit' => 50, 'default' => '', 'comment' => '平台名称(grab/lineman/foodpanda等)'])
                ->addColumn('campaign_id', 'string', ['limit' => 100, 'default' => '', 'comment' => '活动ID'])
                ->addColumn('campaign_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '活动名称'])
                ->addColumn('campaign_name_for_mex', 'string', ['limit' => 255, 'default' => '', 'comment' => '商户提供的活动名称'])
                ->addColumn('campaign_level', 'string', ['limit' => 50, 'default' => '', 'comment' => '活动级别(order/item)'])
                ->addColumn('campaign_type', 'string', ['limit' => 50, 'default' => '', 'comment' => '活动类型(discount/bundle/free_item)'])
                ->addColumn('usage_count', 'integer', ['default' => 0, 'signed' => false, 'comment' => '活动使用次数'])
                ->addColumn('mex_funded_ratio', 'integer', ['default' => 0, 'signed' => false, 'comment' => '商户资金占比(%)'])
                ->addColumn('deducted_amount', 'biginteger', ['default' => 0, 'comment' => '折扣金额(分)'])
                ->addColumn('deducted_part', 'string', ['limit' => 50, 'default' => '', 'comment' => '折扣部分(subtotal/delivery_fee)'])
                ->addColumn('applied_item_ids', 'text', ['null' => true, 'comment' => '应用的商品ID列表(JSON数组)'])
                ->addColumn('free_item_id', 'string', ['limit' => 100, 'default' => '', 'comment' => '赠品ID'])
                ->addColumn('free_item_name', 'string', ['limit' => 255, 'default' => '', 'comment' => '赠品名称'])
                ->addColumn('free_item_quantity', 'integer', ['default' => 0, 'signed' => false, 'comment' => '赠品数量'])
                ->addColumn('free_item_price', 'biginteger', ['default' => 0, 'comment' => '赠品价格(分)'])
                ->addColumn('create_time', 'biginteger', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'biginteger', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'biginteger', ['default' => 0, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'idx_uuid'])
                ->addIndex(['takeout_order_uuid'], ['name' => 'idx_takeout_order_uuid'])
                ->addIndex(['campaign_id'], ['name' => 'idx_campaign_id'])
                ->addIndex(['delete_time'], ['name' => 'idx_delete_time'])
                ->create();
        }
    }
}

