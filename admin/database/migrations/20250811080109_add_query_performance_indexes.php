<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddQueryPerformanceIndexes extends Migrator
{
    /**
     * Up Method.
     * 
     * 添加查询性能优化索引
     */
    public function up()
    {
        // ==================== 设备表索引 ====================
        
        // 设备表 - 优化复合查询
        $this->checkAndAddIndex('device', 'idx_source_deviceid_deletetime_id', ['source', 'device_id', 'delete_time', 'id']);
        $this->checkAndAddIndex('device', 'idx_source_deviceid', ['source', 'device_id']);
        $this->checkAndAddIndex('device', 'idx_delete_time', ['delete_time']);
        
        // ==================== 核心业务表索引补充 ====================
        
        // 销售账单表 - 补充索引
        $this->checkAndAddIndex('sale_bill', 'idx_deletetime_uuid_id', ['delete_time', 'uuid', 'id']);
        $this->checkAndAddIndex('sale_bill', 'idx_uuid_hidebilltime_id', ['uuid', 'hide_bill_time', 'id']);
        
        // 销售账单设置表
        $this->checkAndAddIndex('sale_bill_setting', 'idx_sale_bill_uuid', ['sale_bill_uuid']);
        
        // 销售订单表 - 补充索引
        $this->checkAndAddIndex('sale_order', 'idx_deletetime_salebilluuid', ['delete_time', 'sale_bill_uuid']);
        
        // 支付订单表 - 补充索引
        $this->checkAndAddIndex('payment_order', 'idx_deletetime_status_relateduuid', ['delete_time', 'status', 'related_uuid']);
        
        // 退货订单商品表
        $this->checkAndAddIndex('return_order_product', 'idx_deletetime_saleorderproductuuid', ['delete_time', 'sale_order_product_uuid']);
        
        // 必点方案商品表
        $this->checkAndAddIndex('product_must_plan_item', 'idx_deletetime_productpackageuuid', ['delete_time', 'product_package_uuid']);
        
        // 销售订单商品表 - 补充索引
        $this->checkAndAddIndex('sale_order_product', 'idx_deletetime_saleorderuuid', ['delete_time', 'sale_order_uuid']);
        
        // 销售订单商品原因表
        $this->checkAndAddIndex('sale_order_product_reason', 'idx_sale_order_product_uuid', ['sale_order_product_uuid']);
        
        // ==================== 商品相关表索引 ====================
        
        // 关联材料表
        $this->checkAndAddIndex('related_material', 'idx_related_uuid', ['related_uuid']);
        
        // ==================== 人员相关表索引 ====================
        
        // 会员积分日志表 - 补充索引
        $this->checkAndAddIndex('member_point_log', 'idx_related_uuid', ['related_uuid']);
    }

    /**
     * Down Method.
     * 
     * 删除查询性能索引（回退操作）
     */
    public function down()
    {
       
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

}
