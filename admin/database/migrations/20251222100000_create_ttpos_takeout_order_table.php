<?php

use think\migration\Migrator;

class CreateTtposTakeoutOrderTable extends Migrator
{
    /**
     * 创建外卖订单表（多平台）
     * - ttpos_takeout_order: 外卖订单主表，支持 Grab、Foodpanda、Lineman 等多平台
     */
    public function change()
    {
        // 检查表是否已存在
        if (!$this->hasTable('takeout_order')) {
            $table = $this->table('takeout_order', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '外卖订单表(多平台)',
                'id' => false,
                'primary_key' => ['id']
            ]);

            $table
                // 基础字段
                ->addColumn('id', 'biginteger', ['signed' => false, 'identity' => true, 'comment' => '主键ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一标识'])
                ->addColumn('takeout_order_uuid', 'string', ['limit' => 255, 'default' => '', 'comment' => 'rpc takeout 订单ID'])
                
                // 平台信息
                ->addColumn('platform', 'string', ['limit' => 20, 'default' => '', 'comment' => '外卖平台: grab,foodpanda,lineman,etc'])
                ->addColumn('platform_order_id', 'string', ['limit' => 255, 'default' => '', 'comment' => '平台订单ID (Grab: orderID)'])
                ->addColumn('short_order_number', 'string', ['limit' => 50, 'default' => '', 'comment' => '短订单号(用于展示) (Grab: GF-123)'])
                ->addColumn('merchant_id', 'string', ['limit' => 100, 'default' => '', 'comment' => '商户ID (Grab: merchantID)'])
                ->addColumn('partner_merchant_id', 'string', ['limit' => 100, 'default' => '', 'comment' => '合作伙伴商户ID (Grab: partnerMerchantID)'])
                
                // 订单状态
                ->addColumn('order_state', 'integer', ['limit' => 4, 'signed' => false, 'default' => 1, 'comment' => '订单状态: 0=待接单,1=已接单配餐中, 2=待骑手接单, 3=骑手配送中, 4=已完成, 5=已拒单'])
                ->addColumn('is_abnormal', 'integer', ['limit' => 4, 'signed' => false, 'default' => 0, 'comment' => '是否异常: 0=正常,1=异常'])
                ->addColumn('abnormal_detail', 'text', ['null' => true, 'comment' => '异常详情(JSON)'])
                ->addColumn('stock_status', 'integer', ['limit' => 4, 'signed' => false, 'default' => 1, 'comment' => '库存状态: 1=充足,2=不足'])
                
                // 价格信息（单位：分）
                ->addColumn('subtotal', 'biginteger', ['signed' => true, 'default' => 0, 'comment' => '小计金额 (price.subtotal)'])
                ->addColumn('delivery_fee', 'biginteger', ['signed' => true, 'default' => 0, 'comment' => '配送费 (price.deliveryFee)'])
                ->addColumn('small_order_fee', 'biginteger', ['signed' => true, 'default' => 0, 'comment' => '小单费用 (price.smallOrderFee)'])
                ->addColumn('eater_payment', 'biginteger', ['signed' => true, 'default' => 0, 'comment' => '顾客实付 (price.eaterPayment)'])
                ->addColumn('platform_discount', 'biginteger', ['signed' => true, 'default' => 0, 'comment' => '平台优惠 (price.grabFundPromo)'])
                ->addColumn('merchant_discount', 'biginteger', ['signed' => true, 'default' => 0, 'comment' => '商户优惠 (price.merchantFundPromo)'])
                ->addColumn('basket_promo', 'biginteger', ['signed' => true, 'default' => 0, 'comment' => '购物车优惠 (price.basketPromo)'])
                ->addColumn('tax', 'biginteger', ['signed' => true, 'default' => 0, 'comment' => '税费 (price.tax)'])
                ->addColumn('merchant_charge_fee', 'biginteger', ['signed' => true, 'default' => 0, 'comment' => '商户服务费 (price.merchantChargeFee)'])
                
                // 货币信息
                ->addColumn('currency_code', 'string', ['limit' => 10, 'default' => '', 'comment' => '货币代码(THB,VND等)'])
                ->addColumn('currency_symbol', 'string', ['limit' => 10, 'default' => '', 'comment' => '货币符号(฿,$等)'])
                ->addColumn('currency_exponent', 'integer', ['limit' => 4, 'signed' => false, 'default' => 2, 'comment' => '货币指数'])
                
                // 支付信息
                ->addColumn('payment_type', 'string', ['limit' => 20, 'default' => '', 'comment' => '支付方式: CASH,ONLINE'])
                
                // 订单时间
                ->addColumn('order_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '下单时间 (orderTime)'])
                ->addColumn('submit_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '提交时间 (submitTime)'])
                ->addColumn('scheduled_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '预定时间 (scheduledTime)'])
                ->addColumn('accepted_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '接单时间'])
                ->addColumn('completed_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '完成时间 (completeTime)'])
                ->addColumn('rejected_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '拒单时间'])
                ->addColumn('estimated_ready_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '预计完成时间 (estimatedOrderReadyTime)'])
                ->addColumn('max_ready_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '最大完成时间 (maxOrderReadyTime)'])
                
                // 其他通用信息
                ->addColumn('cutlery', 'integer', ['limit' => 4, 'signed' => false, 'default' => 0, 'comment' => '是否需要餐具: 0=否,1=是'])
                ->addColumn('order_type', 'string', ['limit' => 50, 'default' => '', 'comment' => '订单类型 (featureFlags.orderType): DeliveredByGrab,Pickup,DineIn'])
                ->addColumn('order_accepted_type', 'string', ['limit' => 20, 'default' => '', 'comment' => '接单类型 (featureFlags.orderAcceptedType): AUTO,MANUAL'])
                ->addColumn('is_mex_edit_order', 'integer', ['limit' => 4, 'signed' => false, 'default' => 0, 'comment' => '是否商户编辑订单 (featureFlags.isMexEditOrder)'])
                ->addColumn('membership_id', 'string', ['limit' => 50, 'default' => '', 'comment' => '会员ID (membershipID)'])
                ->addColumn('driver_eta', 'integer', ['signed' => false, 'default' => 0, 'comment' => '司机预计到达时间 (driverETA)'])
                
                // 平台特定数据（JSON 格式）
                ->addColumn('platform_data', 'text', ['limit' => \Phinx\Db\Adapter\MysqlAdapter::TEXT_MEDIUM, 'null' => true, 'comment' => '平台特定字段(JSON): Grab的partner_merchant_id等'])
                
                // 完整原始数据（JSON 格式）
                ->addColumn('raw_data', 'text', ['limit' => \Phinx\Db\Adapter\MysqlAdapter::TEXT_MEDIUM, 'null' => true, 'comment' => '平台原始订单数据(JSON)'])
                
                // 操作信息
                ->addColumn('accepted_by', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '接单人UUID'])
                ->addColumn('rejected_by', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '拒单人UUID'])
                ->addColumn('reject_reason_code', 'string', ['limit' => 50, 'default' => '', 'comment' => '拒单原因代码'])
                ->addColumn('reject_reason', 'string', ['limit' => 255, 'default' => '', 'comment' => '拒单原因'])
                
                // 标准字段
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])
                
                // 索引
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex(['platform', 'platform_order_id', 'delete_time'], ['unique' => true, 'name' => 'uk_platform_order'])
                ->addIndex(['platform', 'delete_time'], ['name' => 'idx_platform'])
                ->addIndex(['order_state', 'delete_time'], ['name' => 'idx_order_state'])
                ->addIndex(['order_time', 'delete_time'], ['name' => 'idx_order_time'])
                ->addIndex(['short_order_number', 'delete_time'], ['name' => 'idx_short_order_number'])
                
                ->create();
        }

        // 检查表是否已存在
        if (!$this->hasTable('takeout_order_item')) {
            $table = $this->table('takeout_order_item', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '外卖订单商品表(多平台)',
                'id' => false,
                'primary_key' => ['id']
            ]);

            $table
                // 基础字段
                ->addColumn('id', 'biginteger', ['signed' => false, 'identity' => true, 'comment' => '主键ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一标识'])
                ->addColumn('takeout_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '外卖订单UUID'])
                ->addColumn('platform', 'string', ['limit' => 20, 'default' => '', 'comment' => '外卖平台: grab,foodpanda,lineman,etc'])
                
                // 平台商品信息
                ->addColumn('platform_item_id', 'string', ['limit' => 100, 'default' => '', 'comment' => '平台商品ID (Grab: TTPOS-ITEM-{uuid})'])
                ->addColumn('item_name', 'text', ['null' => true, 'comment' => '商品名称'])
                
                // TTPOS 商品信息（从 platform_item_id 解析）
                ->addColumn('ttpos_product_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'TTPOS商品UUID (从TTPOS-ITEM-前缀提取)'])
                
                // 商品数量和价格
                ->addColumn('quantity', 'integer', ['signed' => true, 'default' => 0, 'comment' => '数量'])
                ->addColumn('price', 'decimal', ['precision' => 20, 'scale' => 4, 'default' => '0.0000', 'comment' => '单价(元,4位小数)'])
                ->addColumn('tax', 'decimal', ['precision' => 20, 'scale' => 4, 'default' => '0.0000', 'comment' => '税费(元,4位小数)'])
                ->addColumn('specifications', 'string', ['limit' => 500, 'default' => '', 'comment' => '规格说明'])
                
                // 关联状态
                ->addColumn('is_mapped', 'integer', ['limit' => 4, 'signed' => false, 'default' => 0, 'comment' => '是否已关联: 0=无TTPOS前缀(异常),1=有TTPOS前缀(正常)'])
                
                // 标准字段
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])
                
                // 索引
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex(['takeout_order_uuid', 'delete_time'], ['name' => 'idx_takeout_order_uuid'])
                ->addIndex(['platform', 'platform_item_id', 'delete_time'], ['name' => 'idx_platform_item'])
                
                ->create();
        }

        // 检查表是否存在
        if (!$this->hasTable('takeout_order_item_modifier')) {
            $table = $this->table('takeout_order_item_modifier', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_general_ci',
                'comment' => '外卖订单商品修饰符表(多平台)',
                'id' => false,
                'primary_key' => ['id']
            ]);
    
            $table->addColumn('id', 'biginteger', ['signed' => false, 'identity' => true, 'comment' => '主键ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一标识'])
                ->addColumn('takeout_order_item_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '订单商品UUID'])
                ->addColumn('platform', 'string', ['limit' => 50, 'default' => '', 'comment' => '平台: grab,foodpanda,lineman'])
                ->addColumn('platform_modifier_id', 'string', ['limit' => 255, 'default' => '', 'comment' => '平台修饰符ID'])
                ->addColumn('modifier_name', 'text', ['null' => true, 'comment' => '修饰符名称'])
                ->addColumn('ttpos_modifier_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'TTPOS修饰符UUID(关联后：规格/加料/属性值的UUID/套餐商品组UUID)'])
                ->addColumn('ttpos_modifier_type', 'string', ['limit' => 20, 'default' => '', 'comment' => 'TTPOS修饰符类型: flavor=规格, sauce=加料, attr=属性'])
                ->addColumn('quantity', 'integer', ['signed' => false, 'default' => 1, 'comment' => '数量'])
                ->addColumn('price', 'decimal', ['precision' => 20, 'scale' => 4, 'default' => '0.0000', 'comment' => '价格(元,4位小数)'])
                ->addColumn('tax', 'decimal', ['precision' => 20, 'scale' => 4, 'default' => '0.0000', 'comment' => '税费(元,4位小数)'])
                ->addColumn('is_mapped', 'integer', ['limit' => \Phinx\Db\Adapter\MysqlAdapter::INT_TINY, 'default' => 0, 'comment' => '是否已映射: 0=未映射,1=已映射'])
                ->addColumn('create_time', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'idx_uuid'])
                ->addIndex(['takeout_order_item_uuid'], ['name' => 'idx_order_item_uuid'])
                ->addIndex(['platform', 'platform_modifier_id'], ['name' => 'idx_platform_modifier'])
                ->addIndex(['delete_time'], ['name' => 'idx_delete_time'])
                ->create();
        }

        // 检查表是否已存在
        if (!$this->hasTable('takeout_settings')) {
            $table = $this->table('takeout_settings', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '外卖平台配置表(多平台)',
                'id' => false,
                'primary_key' => ['id']
            ]);

            $table
                // 基础字段
                ->addColumn('id', 'biginteger', ['signed' => false, 'identity' => true, 'comment' => '主键ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一标识'])
                ->addColumn('platform', 'string', ['limit' => 20, 'default' => '', 'comment' => '外卖平台: grab,foodpanda,lineman,etc'])
                
                // 基础配置
                ->addColumn('is_enabled', 'integer', ['limit' => 4, 'signed' => false, 'default' => 0, 'comment' => '是否启用: 0=关闭,1=开启'])
                
                // 自动接单配置
                ->addColumn('auto_accept', 'integer', ['limit' => 4, 'signed' => false, 'default' => 0, 'comment' => '自动接单开关: 0=关闭,1=开启'])
                ->addColumn('max_amount', 'biginteger', ['signed' => true, 'default' => 0, 'comment' => '自动接单金额上限(分)'])
                
                // 平台特定配置（JSON 格式）
                ->addColumn('platform_config', 'text', ['null' => true, 'comment' => '平台特定配置(JSON)'])
                
                // 标准字段
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])
                
                // 索引
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex(['platform', 'delete_time'], ['unique' => true, 'name' => 'uk_platform'])
                ->addIndex(['platform', 'delete_time'], ['name' => 'idx_platform'])
                
                ->create();
        }
    }
}

