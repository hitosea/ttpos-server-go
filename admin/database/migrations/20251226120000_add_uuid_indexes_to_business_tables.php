<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddUuidIndexesToBusinessTables extends Migrator
{
    /**
     * Up Method.
     * 
     * 为主要业务表的所有 uuid 字段添加索引，优化查询性能
     */
    public function up()
    {
        // ==================== 销售账单表 (sale_bill) ====================
        $this->checkAndAddIndex('sale_bill', 'idx_consumer_uuid', ['consumer_uuid']);
        $this->checkAndAddIndex('sale_bill', 'idx_cashier_uuid', ['cashier_uuid']);
        $this->checkAndAddIndex('sale_bill', 'idx_member_sale_order_uuid', ['member_sale_order_uuid']);
        $this->checkAndAddIndex('sale_bill', 'idx_batch_tag_uuid', ['batch_tag_uuid']);

        // ==================== 销售订单表 (sale_order) ====================
        $this->checkAndAddIndex('sale_order', 'idx_full_reduction_activity_uuid', ['full_reduction_activity_uuid']);
        $this->checkAndAddIndex('sale_order', 'idx_consumer_uuid', ['consumer_uuid']);
        $this->checkAndAddIndex('sale_order', 'idx_cashier_uuid', ['cashier_uuid']);
        $this->checkAndAddIndex('sale_order', 'idx_staff_shift_log_uuid', ['staff_shift_log_uuid']);

        // ==================== 销售订单原料表 (sale_order_material) ====================
        $this->checkAndAddIndex('sale_order_material', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->checkAndAddIndex('sale_order_material', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->checkAndAddIndex('sale_order_material', 'idx_material_uuid', ['material_uuid']);
        $this->checkAndAddIndex('sale_order_material', 'idx_warehouse_uuid', ['warehouse_uuid']);
        $this->checkAndAddIndex('sale_order_material', 'idx_staff_shift_log_uuid', ['staff_shift_log_uuid']);

        // ==================== 销售订单优惠券表 (sale_order_coupon) ====================
        $this->checkAndAddIndex('sale_order_coupon', 'idx_member_coupon_uuid', ['member_coupon_uuid']);
        $this->checkAndAddIndex('sale_order_coupon', 'idx_marketing_coupon_uuid', ['marketing_coupon_uuid']);

        // ==================== 会员销售订单表 (member_sale_order) ====================
        $this->checkAndAddIndex('member_sale_order', 'idx_member_uuid', ['member_uuid']);
        $this->checkAndAddIndex('member_sale_order', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->checkAndAddIndex('member_sale_order', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->checkAndAddIndex('member_sale_order', 'idx_payment_method_uuid', ['payment_method_uuid']);
        $this->checkAndAddIndex('member_sale_order', 'idx_member_address_uuid', ['member_address_uuid']);

        // ==================== 支付订单表 (payment_order) ====================
        $this->checkAndAddIndex('payment_order', 'idx_payment_method_uuid', ['payment_method_uuid']);

        // ==================== 连连支付订单表 (ll_payment_order) ====================
        $this->checkAndAddIndex('ll_payment_order', 'idx_payment_method_uuid', ['payment_method_uuid']);
        $this->checkAndAddIndex('ll_payment_order', 'idx_member_sale_order_uuid', ['member_sale_order_uuid']);

        // ==================== 销售订单商品表 (sale_order_product) ====================
        $this->checkAndAddIndex('sale_order_product', 'idx_production_order_uuid', ['production_order_uuid']);
        $this->checkAndAddIndex('sale_order_product', 'idx_must_plan_uuid', ['must_plan_uuid']);
        $this->checkAndAddIndex('sale_order_product', 'idx_desk_uuid', ['desk_uuid']);
        $this->checkAndAddIndex('sale_order_product', 'idx_h5_order_uuid', ['h5_order_uuid']);
        $this->checkAndAddIndex('sale_order_product', 'idx_h5_order_product_uuid', ['h5_order_product_uuid']);
        $this->checkAndAddIndex('sale_order_product', 'idx_package_uuid', ['package_uuid']);
        $this->checkAndAddIndex('sale_order_product', 'idx_batch_tag_uuid', ['batch_tag_uuid']);

        // ==================== 销售订单商品原因表 (sale_order_product_reason) ====================
        $this->checkAndAddIndex('sale_order_product_reason', 'idx_return_food_reason_uuid', ['return_food_reason_uuid']);
        $this->checkAndAddIndex('sale_order_product_reason', 'idx_free_reason_uuid', ['free_reason_uuid']);
        $this->checkAndAddIndex('sale_order_product_reason', 'idx_gift_reason_uuid', ['gift_reason_uuid']);

        // ==================== H5订单表 (h5_order) ====================
        $this->checkAndAddIndex('h5_order', 'idx_staff_uuid', ['staff_uuid']);
        $this->checkAndAddIndex('h5_order', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->checkAndAddIndex('h5_order', 'idx_sale_bill_uuid', ['sale_bill_uuid']);

        // ==================== H5订单商品表 (h5_order_product) ====================
        $this->checkAndAddIndex('h5_order_product', 'idx_sale_bill_uuid', ['sale_bill_uuid']);

        // ==================== 销售订单商品BOM表 (sale_order_product_bom) ====================
        $this->checkAndAddIndex('sale_order_product_bom', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->checkAndAddIndex('sale_order_product_bom', 'idx_product_bom_uuid', ['product_bom_uuid']);

        // ==================== 销售订单商品属性表 (sale_order_product_attribute) ====================
        $this->checkAndAddIndex('sale_order_product_attribute', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->checkAndAddIndex('sale_order_product_attribute', 'idx_product_attribute_uuid', ['product_attribute_uuid']);
        $this->checkAndAddIndex('sale_order_product_attribute', 'idx_product_package_attribute_uuid', ['product_package_attribute_uuid']);

        // ==================== 销售订单优惠策略表 (sale_order_discount_strategy) ====================
        $this->checkAndAddIndex('sale_order_discount_strategy', 'idx_sale_order_uuid', ['sale_order_uuid']);

        // ==================== 生产订单表 (production_order) ====================
        $this->checkAndAddIndex('production_order', 'idx_desk_uuid', ['desk_uuid']);
        $this->checkAndAddIndex('production_order', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->checkAndAddIndex('production_order', 'idx_sale_bill_uuid', ['sale_bill_uuid']);

        // ==================== 生产订单商品表 (production_order_product) ====================
        $this->checkAndAddIndex('production_order_product', 'idx_product_bom_uuid', ['product_bom_uuid']);
        $this->checkAndAddIndex('production_order_product', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->checkAndAddIndex('production_order_product', 'idx_product_package_uuid', ['product_package_uuid']);
        $this->checkAndAddIndex('production_order_product', 'idx_production_order_uuid', ['production_order_uuid']);
        $this->checkAndAddIndex('production_order_product', 'idx_first_category_uuid', ['first_category_uuid']);
        $this->checkAndAddIndex('production_order_product', 'idx_batch_tag_uuid', ['batch_tag_uuid']);

        // ==================== 生产订单原料表 (production_order_material) ====================
        $this->checkAndAddIndex('production_order_material', 'idx_material_uuid', ['material_uuid']);
        $this->checkAndAddIndex('production_order_material', 'idx_production_order_product_uuid', ['production_order_product_uuid']);
        $this->checkAndAddIndex('production_order_material', 'idx_sale_order_product_uuid', ['sale_order_product_uuid']);

        // ==================== 桌台表 (desk) ====================
        $this->checkAndAddIndex('desk', 'idx_region_uuid', ['region_uuid']);
        $this->checkAndAddIndex('desk', 'idx_type_uuid', ['type_uuid']);
        $this->checkAndAddIndex('desk', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->checkAndAddIndex('desk', 'idx_device_uuid', ['device_uuid']);

        // ==================== 销售订单操作记录表 (sale_order_operation_record) ====================
        $this->checkAndAddIndex('sale_order_operation_record', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->checkAndAddIndex('sale_order_operation_record', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->checkAndAddIndex('sale_order_operation_record', 'idx_h5_order_uuid', ['h5_order_uuid']);
        $this->checkAndAddIndex('sale_order_operation_record', 'idx_operator_uuid', ['operator_uuid']);
        $this->checkAndAddIndex('sale_order_operation_record', 'idx_member_sale_order_uuid', ['member_sale_order_uuid']);
        $this->checkAndAddIndex('sale_order_operation_record', 'idx_member_uuid', ['member_uuid']);

        // ==================== 会员表 (member) ====================
        $this->checkAndAddIndex('member', 'idx_member_level_uuid', ['member_level_uuid']);
        $this->checkAndAddIndex('member', 'idx_member_card_uuid', ['member_card_uuid']);
        $this->checkAndAddIndex('member', 'idx_referrer_uuid', ['referrer_uuid']);
        $this->checkAndAddIndex('member', 'idx_activity_uuid', ['activity_uuid']);

        // ==================== 会员等级日志表 (member_level_log) ====================
        $this->checkAndAddIndex('member_level_log', 'idx_member_uuid', ['member_uuid']);

        // ==================== 会员卡表 (member_card) ====================
        $this->checkAndAddIndex('member_card', 'idx_card_type_uuid', ['card_type_uuid']);
        $this->checkAndAddIndex('member_card', 'idx_member_uuid', ['member_uuid']);

        // ==================== 会员卡日志表 (member_card_log) ====================
        $this->checkAndAddIndex('member_card_log', 'idx_member_card_type_uuid', ['member_card_type_uuid']);
        $this->checkAndAddIndex('member_card_log', 'idx_member_uuid', ['member_uuid']);

        // ==================== 会员余额日志表 (member_balance_log) ====================
        $this->checkAndAddIndex('member_balance_log', 'idx_sale_order_uuid', ['sale_order_uuid']);

        // ==================== 会员充值订单表 (member_recharge_order) ====================
        $this->checkAndAddIndex('member_recharge_order', 'idx_staff_uuid', ['staff_uuid']);

        // ==================== 会员充值订单操作日志表 (member_recharge_order_operation_log) ====================
        $this->checkAndAddIndex('member_recharge_order_operation_log', 'idx_recharge_order_uuid', ['recharge_order_uuid']);

        // ==================== 会员充值订单异常记录表 (member_recharge_order_abnormal_record) ====================
        $this->checkAndAddIndex('member_recharge_order_abnormal_record', 'idx_recharge_order_uuid', ['recharge_order_uuid']);
        $this->checkAndAddIndex('member_recharge_order_abnormal_record', 'idx_cashier_uuid', ['cashier_uuid']);

        // ==================== 原料表 (material) ====================
        $this->checkAndAddIndex('material', 'idx_category_uuid', ['category_uuid']);
        $this->checkAndAddIndex('material', 'idx_supplier_uuid', ['supplier_uuid']);
        $this->checkAndAddIndex('material', 'idx_image_uuid', ['image_uuid']);
        $this->checkAndAddIndex('material', 'idx_unit_uuid', ['unit_uuid']);
        $this->checkAndAddIndex('material', 'idx_purchase_unit_uuid', ['purchase_unit_uuid']);
        $this->checkAndAddIndex('material', 'idx_cost_unit_uuid', ['cost_unit_uuid']);
        $this->checkAndAddIndex('material', 'idx_headquarter_uuid', ['headquarter_uuid']);
        $this->checkAndAddIndex('material', 'idx_warehouse_uuid', ['warehouse_uuid']);

        // ==================== 原料分类表 (material_category) ====================
        $this->checkAndAddIndex('material_category', 'idx_headquarter_uuid', ['headquarter_uuid']);

        // ==================== 仓库表 (warehouse) ====================
        $this->checkAndAddIndex('warehouse', 'idx_headquarter_uuid', ['headquarter_uuid']);
    }

    /**
     * Down Method.
     * 
     * 删除索引（回退操作）
     */
    public function down()
    {
        // 回退操作：删除所有添加的索引
        // 注意：由于索引较多，建议在生产环境回退时谨慎操作
        
        // 销售账单表
        $this->removeIndexIfExists('sale_bill', 'idx_consumer_uuid', ['consumer_uuid']);
        $this->removeIndexIfExists('sale_bill', 'idx_cashier_uuid', ['cashier_uuid']);
        $this->removeIndexIfExists('sale_bill', 'idx_member_sale_order_uuid', ['member_sale_order_uuid']);
        $this->removeIndexIfExists('sale_bill', 'idx_batch_tag_uuid', ['batch_tag_uuid']);

        // 销售订单表
        $this->removeIndexIfExists('sale_order', 'idx_full_reduction_activity_uuid', ['full_reduction_activity_uuid']);
        $this->removeIndexIfExists('sale_order', 'idx_consumer_uuid', ['consumer_uuid']);
        $this->removeIndexIfExists('sale_order', 'idx_cashier_uuid', ['cashier_uuid']);
        $this->removeIndexIfExists('sale_order', 'idx_staff_shift_log_uuid', ['staff_shift_log_uuid']);

        // 销售订单原料表
        $this->removeIndexIfExists('sale_order_material', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->removeIndexIfExists('sale_order_material', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->removeIndexIfExists('sale_order_material', 'idx_material_uuid', ['material_uuid']);
        $this->removeIndexIfExists('sale_order_material', 'idx_warehouse_uuid', ['warehouse_uuid']);
        $this->removeIndexIfExists('sale_order_material', 'idx_staff_shift_log_uuid', ['staff_shift_log_uuid']);

        // 销售订单优惠券表
        $this->removeIndexIfExists('sale_order_coupon', 'idx_member_coupon_uuid', ['member_coupon_uuid']);
        $this->removeIndexIfExists('sale_order_coupon', 'idx_marketing_coupon_uuid', ['marketing_coupon_uuid']);

        // 会员销售订单表
        $this->removeIndexIfExists('member_sale_order', 'idx_member_uuid', ['member_uuid']);
        $this->removeIndexIfExists('member_sale_order', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->removeIndexIfExists('member_sale_order', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->removeIndexIfExists('member_sale_order', 'idx_payment_method_uuid', ['payment_method_uuid']);
        $this->removeIndexIfExists('member_sale_order', 'idx_member_address_uuid', ['member_address_uuid']);

        // 支付订单表
        $this->removeIndexIfExists('payment_order', 'idx_payment_method_uuid', ['payment_method_uuid']);

        // 连连支付订单表
        $this->removeIndexIfExists('ll_payment_order', 'idx_payment_method_uuid', ['payment_method_uuid']);
        $this->removeIndexIfExists('ll_payment_order', 'idx_member_sale_order_uuid', ['member_sale_order_uuid']);

        // 销售订单商品表
        $this->removeIndexIfExists('sale_order_product', 'idx_production_order_uuid', ['production_order_uuid']);
        $this->removeIndexIfExists('sale_order_product', 'idx_must_plan_uuid', ['must_plan_uuid']);
        $this->removeIndexIfExists('sale_order_product', 'idx_desk_uuid', ['desk_uuid']);
        $this->removeIndexIfExists('sale_order_product', 'idx_h5_order_uuid', ['h5_order_uuid']);
        $this->removeIndexIfExists('sale_order_product', 'idx_h5_order_product_uuid', ['h5_order_product_uuid']);
        $this->removeIndexIfExists('sale_order_product', 'idx_package_uuid', ['package_uuid']);
        $this->removeIndexIfExists('sale_order_product', 'idx_batch_tag_uuid', ['batch_tag_uuid']);

        // 销售订单商品原因表
        $this->removeIndexIfExists('sale_order_product_reason', 'idx_return_food_reason_uuid', ['return_food_reason_uuid']);
        $this->removeIndexIfExists('sale_order_product_reason', 'idx_free_reason_uuid', ['free_reason_uuid']);
        $this->removeIndexIfExists('sale_order_product_reason', 'idx_gift_reason_uuid', ['gift_reason_uuid']);

        // H5订单表
        $this->removeIndexIfExists('h5_order', 'idx_staff_uuid', ['staff_uuid']);
        $this->removeIndexIfExists('h5_order', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->removeIndexIfExists('h5_order', 'idx_sale_bill_uuid', ['sale_bill_uuid']);

        // H5订单商品表
        $this->removeIndexIfExists('h5_order_product', 'idx_sale_bill_uuid', ['sale_bill_uuid']);

        // 销售订单商品BOM表
        $this->removeIndexIfExists('sale_order_product_bom', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->removeIndexIfExists('sale_order_product_bom', 'idx_product_bom_uuid', ['product_bom_uuid']);

        // 销售订单商品属性表
        $this->removeIndexIfExists('sale_order_product_attribute', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->removeIndexIfExists('sale_order_product_attribute', 'idx_product_attribute_uuid', ['product_attribute_uuid']);
        $this->removeIndexIfExists('sale_order_product_attribute', 'idx_product_package_attribute_uuid', ['product_package_attribute_uuid']);

        // 销售订单优惠策略表
        $this->removeIndexIfExists('sale_order_discount_strategy', 'idx_sale_order_uuid', ['sale_order_uuid']);

        // 生产订单表
        $this->removeIndexIfExists('production_order', 'idx_desk_uuid', ['desk_uuid']);
        $this->removeIndexIfExists('production_order', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->removeIndexIfExists('production_order', 'idx_sale_bill_uuid', ['sale_bill_uuid']);

        // 生产订单商品表
        $this->removeIndexIfExists('production_order_product', 'idx_product_bom_uuid', ['product_bom_uuid']);
        $this->removeIndexIfExists('production_order_product', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->removeIndexIfExists('production_order_product', 'idx_product_package_uuid', ['product_package_uuid']);
        $this->removeIndexIfExists('production_order_product', 'idx_production_order_uuid', ['production_order_uuid']);
        $this->removeIndexIfExists('production_order_product', 'idx_first_category_uuid', ['first_category_uuid']);
        $this->removeIndexIfExists('production_order_product', 'idx_batch_tag_uuid', ['batch_tag_uuid']);

        // 生产订单原料表
        $this->removeIndexIfExists('production_order_material', 'idx_material_uuid', ['material_uuid']);
        $this->removeIndexIfExists('production_order_material', 'idx_production_order_product_uuid', ['production_order_product_uuid']);
        $this->removeIndexIfExists('production_order_material', 'idx_sale_order_product_uuid', ['sale_order_product_uuid']);

        // 桌台表
        $this->removeIndexIfExists('desk', 'idx_region_uuid', ['region_uuid']);
        $this->removeIndexIfExists('desk', 'idx_type_uuid', ['type_uuid']);
        $this->removeIndexIfExists('desk', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->removeIndexIfExists('desk', 'idx_device_uuid', ['device_uuid']);

        // 销售订单操作记录表
        $this->removeIndexIfExists('sale_order_operation_record', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        $this->removeIndexIfExists('sale_order_operation_record', 'idx_sale_order_uuid', ['sale_order_uuid']);
        $this->removeIndexIfExists('sale_order_operation_record', 'idx_h5_order_uuid', ['h5_order_uuid']);
        $this->removeIndexIfExists('sale_order_operation_record', 'idx_operator_uuid', ['operator_uuid']);
        $this->removeIndexIfExists('sale_order_operation_record', 'idx_member_sale_order_uuid', ['member_sale_order_uuid']);
        $this->removeIndexIfExists('sale_order_operation_record', 'idx_member_uuid', ['member_uuid']);

        // 会员表
        $this->removeIndexIfExists('member', 'idx_member_level_uuid', ['member_level_uuid']);
        $this->removeIndexIfExists('member', 'idx_member_card_uuid', ['member_card_uuid']);
        $this->removeIndexIfExists('member', 'idx_referrer_uuid', ['referrer_uuid']);
        $this->removeIndexIfExists('member', 'idx_activity_uuid', ['activity_uuid']);

        // 会员等级日志表
        $this->removeIndexIfExists('member_level_log', 'idx_member_uuid', ['member_uuid']);

        // 会员卡表
        $this->removeIndexIfExists('member_card', 'idx_card_type_uuid', ['card_type_uuid']);
        $this->removeIndexIfExists('member_card', 'idx_member_uuid', ['member_uuid']);

        // 会员卡日志表
        $this->removeIndexIfExists('member_card_log', 'idx_member_card_type_uuid', ['member_card_type_uuid']);
        $this->removeIndexIfExists('member_card_log', 'idx_member_uuid', ['member_uuid']);

        // 会员余额日志表
        $this->removeIndexIfExists('member_balance_log', 'idx_sale_order_uuid', ['sale_order_uuid']);

        // 会员充值订单表
        $this->removeIndexIfExists('member_recharge_order', 'idx_staff_uuid', ['staff_uuid']);

        // 会员充值订单操作日志表
        $this->removeIndexIfExists('member_recharge_order_operation_log', 'idx_recharge_order_uuid', ['recharge_order_uuid']);

        // 会员充值订单异常记录表
        $this->removeIndexIfExists('member_recharge_order_abnormal_record', 'idx_recharge_order_uuid', ['recharge_order_uuid']);
        $this->removeIndexIfExists('member_recharge_order_abnormal_record', 'idx_cashier_uuid', ['cashier_uuid']);

        // 原料表
        $this->removeIndexIfExists('material', 'idx_category_uuid', ['category_uuid']);
        $this->removeIndexIfExists('material', 'idx_supplier_uuid', ['supplier_uuid']);
        $this->removeIndexIfExists('material', 'idx_image_uuid', ['image_uuid']);
        $this->removeIndexIfExists('material', 'idx_unit_uuid', ['unit_uuid']);
        $this->removeIndexIfExists('material', 'idx_purchase_unit_uuid', ['purchase_unit_uuid']);
        $this->removeIndexIfExists('material', 'idx_cost_unit_uuid', ['cost_unit_uuid']);
        $this->removeIndexIfExists('material', 'idx_headquarter_uuid', ['headquarter_uuid']);
        $this->removeIndexIfExists('material', 'idx_warehouse_uuid', ['warehouse_uuid']);

        // 原料分类表
        $this->removeIndexIfExists('material_category', 'idx_headquarter_uuid', ['headquarter_uuid']);

        // 仓库表
        $this->removeIndexIfExists('warehouse', 'idx_headquarter_uuid', ['headquarter_uuid']);
    }

    /**
     * 检查并添加索引
     * @param string $tableName 表名（不带 ttpos_ 前缀）
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
            
            // 检查索引是否已存在（使用 hasIndexByName 方法）
            if ($table->hasIndex($indexName)) {
                return;
            }

            // 添加索引
            $table->addIndex($columns, [
                'name' => $indexName,
                'unique' => false
            ])->update();
        } catch (\Exception $e) {
            // 索引已存在或其他错误，忽略
        }
    }

    /**
     * 删除索引（如果存在）
     * @param string $tableName 表名（不带 ttpos_ 前缀）
     * @param string $indexName 索引名
     * @param array $columns 索引字段（用于删除索引）
     */
    protected function removeIndexIfExists($tableName, $indexName, $columns)
    {
        try {
            // 检查表是否存在
            if (!$this->hasTable($tableName)) {
                return;
            }

            $table = $this->table($tableName);
            
            // 检查索引是否存在
            if ($table->hasIndex($indexName)) {
                $table->removeIndex($columns, [
                    'name' => $indexName
                ])->update();
            }
        } catch (\Exception $e) {
            // 忽略错误
        }
    }
}

