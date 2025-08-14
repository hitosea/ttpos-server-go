<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddCriticalPerformanceIndexes extends Migrator
{
    /**
     * Up Method.
     * 
     * 添加数据库索引
     */
    public function up()
    {
        // ==================== 核心业务表索引 ====================
        
        // 销售账单表 - 关键业务查询索引
        $this->checkAndAddIndex('sale_bill', 'idx_device_uuid_status', ['device_uuid', 'status', 'delete_time']);
        $this->checkAndAddIndex('sale_bill', 'idx_desk_uuid_status', ['desk_uuid', 'status', 'delete_time']);
        $this->checkAndAddIndex('sale_bill', 'idx_status_delete_time', ['status', 'delete_time']);
        $this->checkAndAddIndex('sale_bill', 'idx_create_time', ['create_time']);
        
        // 销售订单表 - 订单查询优化
        $this->checkAndAddIndex('sale_order', 'idx_sale_bill_uuid_status', ['sale_bill_uuid', 'status', 'delete_time']);
        $this->checkAndAddIndex('sale_order', 'idx_create_time', ['create_time']);
        $this->checkAndAddIndex('sale_order', 'idx_status_delete_time', ['status', 'delete_time']);
        
        // 销售订单商品表 - 商品查询优化
        $this->checkAndAddIndex('sale_order_product', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->checkAndAddIndex('sale_order_product', 'idx_product_package_uuid', ['product_package_uuid']);
        $this->checkAndAddIndex('sale_order_product', 'idx_status_delete_time', ['status', 'delete_time']);
        $this->checkAndAddIndex('sale_order_product', 'idx_is_accept_order', ['is_accept_order']);
        
        // H5订单表 - 线上订单查询优化
        $this->checkAndAddIndex('h5_order', 'idx_desk_uuid_status', ['desk_uuid', 'status', 'delete_time']);
        $this->checkAndAddIndex('h5_order', 'idx_create_time', ['create_time']);
        $this->checkAndAddIndex('h5_order', 'idx_status_auto_accept', ['status', 'is_auto_accept']);
        
        // H5订单商品表
        $this->checkAndAddIndex('h5_order_product', 'idx_h5_order_uuid_delete', ['h5_order_uuid', 'delete_time']);
        $this->checkAndAddIndex('h5_order_product', 'idx_sale_order_product_uuid', ['sale_order_product_uuid']);
        
        // ==================== 会员相关表索引 ====================
        
        // 会员表 - 会员查询优化
        $this->checkAndAddIndex('member', 'idx_phone', ['phone']);
        $this->checkAndAddIndex('member', 'idx_device_id', ['device_id']); // 游客设备ID查询
        $this->checkAndAddIndex('member', 'idx_is_visitor', ['is_visitor']);
        $this->checkAndAddIndex('member', 'idx_create_time', ['create_time']);
        
        // 会员地址表
        $this->checkAndAddIndex('member_address', 'idx_member_uuid_delete', ['member_uuid', 'delete_time']);
        $this->checkAndAddIndex('member_address', 'idx_is_default', ['is_default']);
        
        // 会员积分日志表
        $this->checkAndAddIndex('member_point_log', 'idx_member_uuid_scene', ['member_uuid', 'scene']);
        $this->checkAndAddIndex('member_point_log', 'idx_related_uuid', ['related_uuid']);
        $this->checkAndAddIndex('member_point_log', 'idx_create_time', ['create_time']);
        
        // 会员余额日志表
        $this->checkAndAddIndex('member_balance_log', 'idx_member_uuid_scene', ['member_uuid', 'scene']);
        $this->checkAndAddIndex('member_balance_log', 'idx_related_uuid', ['related_uuid']);
        $this->checkAndAddIndex('member_balance_log', 'idx_create_time', ['create_time']);
        
        // 会员充值订单表
        $this->checkAndAddIndex('member_recharge_order', 'idx_member_uuid_status', ['member_uuid', 'status']);
        $this->checkAndAddIndex('member_recharge_order', 'idx_create_time', ['create_time']);
        $this->checkAndAddIndex('member_recharge_order', 'idx_status', ['status']);
        
        // ==================== 支付相关表索引 ====================
        
        // 支付订单表
        $this->checkAndAddIndex('payment_order', 'idx_related_uuid', ['related_uuid']);
        $this->checkAndAddIndex('payment_order', 'idx_status', ['status']);
        $this->checkAndAddIndex('payment_order', 'idx_create_time', ['create_time']);
        
        // 销售订单支付表
        $this->checkAndAddIndex('sale_order_payment', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->checkAndAddIndex('sale_order_payment', 'idx_payment_method_uuid', ['payment_method_uuid']);
        
        // ==================== 打印相关表索引 ====================
        
        // 打印机表
        $this->checkAndAddIndex('printer', 'idx_status_delete', ['status', 'delete_time']);
        $this->checkAndAddIndex('printer', 'idx_is_usb', ['is_usb']);
        
        // 打印日志表
        $this->checkAndAddIndex('printer_log', 'idx_printer_uuid', ['printer_uuid']);
        $this->checkAndAddIndex('printer_log', 'idx_status', ['status']);
        $this->checkAndAddIndex('printer_log', 'idx_create_time', ['create_time']);
        
        // ==================== 库存相关表索引 ====================
        
        // 出库表项
        $this->checkAndAddIndex('warehouse_out_form_item', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->checkAndAddIndex('warehouse_out_form_item', 'idx_reduce_stock', ['reduce_stock']);
        
        // 入库表项
        $this->checkAndAddIndex('warehouse_form_item', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->checkAndAddIndex('warehouse_form_item', 'idx_add_stock', ['add_stock']);
        
        // ==================== 营销活动相关表索引 ====================
        
        // 营销活动表
        $this->checkAndAddIndex('marketing_activity', 'idx_status_time', ['delete_time', 'start_time', 'end_time', 'is_invalid']);
        
        // 营销活动记录表
        $this->checkAndAddIndex('marketing_activity_record', 'idx_activity_member', ['activity_uuid', 'member_uuid', 'delete_time']);
        
        // ==================== 其他业务表索引 ====================
        
        // 桌台表
        $this->checkAndAddIndex('desk', 'idx_region_uuid_status', ['region_uuid', 'status']);
        
        // 客户呼叫表
        $this->checkAndAddIndex('customer_call', 'idx_desk_uuid_create', ['desk_uuid', 'create_time']);
        $this->checkAndAddIndex('customer_call', 'idx_status', ['status']);
        
        // 操作记录表
        $this->checkAndAddIndex('sale_order_operation_record', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->checkAndAddIndex('sale_order_operation_record', 'idx_h5_order_uuid', ['h5_order_uuid']);
        
        // 统计表索引
        $this->checkAndAddIndex('statistics_sale', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->checkAndAddIndex('statistics_payment', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->checkAndAddIndex('statistics_product', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->checkAndAddIndex('statistics_customer_type', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->checkAndAddIndex('statistics_delay', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->checkAndAddIndex('statistics_member', 'idx_member_recharge_order_uuid', ['member_recharge_order_uuid']);
        $this->checkAndAddIndex('statistics_member_payment', 'idx_member_recharge_order_uuid', ['member_recharge_order_uuid']);
        
        // 商品相关表
        $this->checkAndAddIndex('product_package', 'idx_category_uuid', ['category_uuid']);
        $this->checkAndAddIndex('product_category', 'idx_parent_uuid', ['parent_uuid']);
        
        // 订单商品属性和配方表
        $this->checkAndAddIndex('sale_order_product_bom', 'idx_sale_order_product_uuid', ['sale_order_product_uuid']);
        $this->checkAndAddIndex('sale_order_product_attribute', 'idx_sale_order_product_uuid', ['sale_order_product_uuid']);
        
        // 生产订单商品表
        $this->checkAndAddIndex('production_order_product', 'idx_sale_order_product_uuid', ['sale_order_product_uuid']);
        $this->checkAndAddIndex('production_order_product', 'idx_status', ['status']);
    }

    /**
     * Down Method.
     * 
     * 删除数据库索引（回退操作）
     */
    public function down()
    {
        // 定义所有需要删除的索引
        $indexesToRemove = [
            // 销售账单表索引
            'sale_bill' => [
                'idx_device_uuid_status', 'idx_desk_uuid_status', 'idx_status_delete_time',
                'idx_create_time'
            ],
            // 销售订单表索引
            'sale_order' => [
                'idx_sale_bill_uuid_status',
                'idx_create_time', 'idx_status_delete_time'
            ],
            // 销售订单商品表索引
            'sale_order_product' => [
                'idx_sale_bill_uuid', 'idx_product_package_uuid', 'idx_status_delete_time',
                'idx_is_accept_order'
            ],
            // H5订单表索引
            'h5_order' => [
                'idx_desk_uuid_status', 'idx_create_time',
                'idx_status_auto_accept'
            ],
            // H5订单商品表索引
            'h5_order_product' => [
                'idx_h5_order_uuid_delete', 'idx_sale_order_product_uuid'
            ],
            // 会员表索引
            'member' => [
                'idx_phone', 'idx_device_id', 'idx_is_visitor',
                'idx_create_time'
            ],
            // 会员地址表索引
            'member_address' => [
                'idx_member_uuid_delete', 'idx_is_default'
            ],
            // 会员积分日志表索引
            'member_point_log' => [
                'idx_member_uuid_scene', 'idx_related_uuid', 'idx_create_time'
            ],
            // 会员余额日志表索引
            'member_balance_log' => [
                'idx_member_uuid_scene', 'idx_related_uuid', 'idx_create_time'
            ],
            // 会员充值订单表索引
            'member_recharge_order' => [
                'idx_member_uuid_status', 'idx_create_time', 'idx_status'
            ],
            // 支付订单表索引
            'payment_order' => [
                'idx_related_uuid', 'idx_status', 'idx_create_time'
            ],
            // 销售订单支付表索引
            'sale_order_payment' => [
                'idx_sale_order_uuid', 'idx_payment_method_uuid'
            ],
            // 打印机表索引
            'printer' => [
                'idx_status_delete', 'idx_is_usb'
            ],
            // 打印日志表索引
            'printer_log' => [
                'idx_printer_uuid', 'idx_status', 'idx_create_time'
            ],
            // 出库表项索引
            'warehouse_out_form_item' => [
                'idx_sale_order_uuid', 'idx_reduce_stock'
            ],
            // 入库表项索引
            'warehouse_form_item' => [
                'idx_sale_bill_uuid', 'idx_add_stock'
            ],
            // 营销活动表索引
            'marketing_activity' => [
                'idx_status_time'
            ],
            // 营销活动记录表索引
            'marketing_activity_record' => [
                'idx_activity_member'
            ],
            // 桌台表索引
            'desk' => [
                'idx_region_uuid_status'
            ],
            // 客户呼叫表索引
            'customer_call' => [
                'idx_desk_uuid_create', 'idx_status'
            ],
            // 操作记录表索引
            'sale_order_operation_record' => [
                'idx_sale_bill_uuid', 'idx_h5_order_uuid'
            ],
            // 统计表索引
            'statistics_sale' => ['idx_sale_bill_uuid'],
            'statistics_payment' => ['idx_sale_bill_uuid'],
            'statistics_product' => ['idx_sale_bill_uuid'],
            'statistics_customer_type' => ['idx_sale_bill_uuid'],
            'statistics_delay' => ['idx_sale_bill_uuid'],
            'statistics_member' => ['idx_member_recharge_order_uuid'],
            'statistics_member_payment' => ['idx_member_recharge_order_uuid'],
            // 商品相关表索引
            'product_package' => [
                'idx_category_uuid'
            ],
            'product_category' => [
                'idx_parent_uuid'
            ],
            // 订单商品属性和配方表索引
            'sale_order_product_bom' => [
                'idx_sale_order_product_uuid'
            ],
            'sale_order_product_attribute' => [
                'idx_sale_order_product_uuid'
            ],
            // 生产订单商品表索引
            'production_order_product' => [
                'idx_sale_order_product_uuid', 'idx_status'
            ]
        ];

        // 删除所有索引
        foreach ($indexesToRemove as $tableName => $indexes) {
            $this->removeIndexes($tableName, $indexes);
        }
    }

    /**
     * 检查并添加索引
     * @param string $tableName 表名
     * @param string $indexName 索引名
     * @param array $columns 索引字段
     */
    protected function checkAndAddIndex($tableName, $indexName, $columns)
    {
        try {
            // 检查表是否存在
            if (!$this->hasTable($tableName)) {
                return;
            }

            $table = $this->table($tableName);
            $table->addIndex($columns, [
                'name' => $indexName,
                'unique' => false
            ])->update();
        } catch (\Exception $e) {
            // 索引已存在或其他错误，忽略
        }
    }

    /**
     * 删除表的多个索引
     * @param string $tableName 表名
     * @param array $indexNames 索引名数组
     */
    protected function removeIndexes($tableName, $indexNames)
    {
        // try {
        //     // 检查表是否存在
        //     if (!$this->hasTable($tableName)) {
        //         return;
        //     }

        //     // 使用原生SQL删除索引，这是最可靠的方法
        //     foreach ($indexNames as $indexName) {
        //         try {
        //             // 直接执行SQL删除索引
        //             $sql = "DROP INDEX `{$indexName}` ON `ttpos_{$tableName}`";
        //             $this->execute($sql);
        //         } catch (\Exception $e) {
        //             // 索引不存在或其他错误，忽略
        //             // echo "删除索引失败: " . $e->getMessage();
        //         }
        //     }
        // } catch (\Exception $e) {
        //     // 表不存在或其他错误，忽略
        // }
    }
}