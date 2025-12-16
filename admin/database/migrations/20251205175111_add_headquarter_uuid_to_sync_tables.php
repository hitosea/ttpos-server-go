<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddHeadquarterUuidToSyncTables extends Migrator
{
    /**
     * 为同步相关表添加 headquarter_uuid 字段
     * 用于标识数据来源：0表示本店创建，>0表示从总部同步
     * Requirement: shop-headquarters-branch-granular-sync-backend
     * Task: DooTask #37462
     */
    public function change()
    {
        // 1. 优惠券表
        $this->addHeadquarterUuidColumn('marketing_coupon');
        
        // 2. 满额减活动表
        $this->addHeadquarterUuidColumn('full_reduction_activity');
        
        // 3. 菜品标签表
        $this->addHeadquarterUuidColumn('product_label');
        
        // 4. 营销活动表
        $this->addHeadquarterUuidColumn('marketing_activity');
        
        // 5. 支付方式表
        $this->addHeadquarterUuidColumn('payment_method');
    }
    
    /**
     * 为指定表添加 headquarter_uuid 字段和索引
     * 
     * @param string $tableName 表名（不含前缀）
     */
    private function addHeadquarterUuidColumn($tableName)
    {
        $table = $this->table($tableName);
        
        // 检查字段是否已存在（幂等性）
        if (!$table->hasColumn('headquarter_uuid')) {
            $table->addColumn(
                'headquarter_uuid',
                'biginteger',
                [
                    'null' => false,
                    'default' => 0,
                    'comment' => '总部uuid，0表示本店创建，>0表示从总部同步',
                    'after' => 'uuid'
                ]
            );
            
            $table->update();
        }
        
        // 添加索引（如果已存在则忽略）
        try {
            $table->addIndex('headquarter_uuid', ['name' => 'idx_headquarter_uuid'])->update();
        } catch (\Exception $e) {
            // 索引已存在或其他错误，忽略
        }
    }
}
