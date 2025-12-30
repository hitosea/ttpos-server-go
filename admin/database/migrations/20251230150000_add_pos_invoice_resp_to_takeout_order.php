<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddPosInvoiceRespToTakeoutOrder extends Migrator
{
    /**
     * 为 ttpos_takeout_order 表添加 ERP POS Invoice 响应字段
     * 用于存储 ERP 同步后的发票信息（JSON格式）
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('takeout_order')) {
            $table = $this->table('takeout_order');

            // 检查字段是否已存在，如果不存在则添加
            if (!$table->hasColumn('erp_pos_invoice_resp')) {
                $table->addColumn('erp_pos_invoice_resp', 'text', [
                    'null' => true,
                    'comment' => 'ERP POS Invoice响应数据(JSON格式)',
                    'after' => 'staff_shift_log_uuid',
                ])->update();
            }

            // 检查字段是否已存在
            if (!$table->hasColumn('staff_shift_log_uuid')) {
                $table->addColumn('staff_shift_log_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '员工班次日志UUID', 'after' => 'accepted_by'])->update();
            }
            
            // 检查索引是否已存在
            if (!$table->hasIndexByName('idx_staff_shift_log_uuid')) {
                $table->addIndex(['staff_shift_log_uuid'], ['name' => 'idx_staff_shift_log_uuid'])->update();
            }
            
            // 修改价格相关字段类型为 decimal(20,4)
            $priceFields = [
                'subtotal' => '小计金额 (price.subtotal)',
                'delivery_fee' => '配送费 (price.deliveryFee)',
                'small_order_fee' => '小单费用 (price.smallOrderFee)',
                'eater_payment' => '顾客实付 (price.eaterPayment)',
                'platform_discount' => '平台优惠 (price.grabFundPromo)',
                'merchant_discount' => '商户优惠 (price.merchantFundPromo)',
                'basket_promo' => '购物车优惠 (price.basketPromo)',
                'tax' => '税费 (price.tax)',
                'merchant_charge_fee' => '商户服务费 (price.merchantChargeFee)',
            ];
            
            foreach ($priceFields as $fieldName => $comment) {
                if ($table->hasColumn($fieldName)) {
                    $table->changeColumn($fieldName, 'decimal', [
                        'precision' => 20,
                        'scale' => 4,
                        'null' => false,
                        'default' => 0.0000,
                        'comment' => $comment,
                    ])->update();
                }
            }
        }
        
        // 1. 处理 takeout_order_item 表
        if ($this->hasTable('takeout_order_item')) {
            $itemTable = $this->table('takeout_order_item');
            if (!$itemTable->hasColumn('ttpos_item_erp_code')) {
                $itemTable->addColumn('ttpos_item_erp_code', 'string', ['limit' => 50, 'default' => '', 'comment' => 'TTPOS商品ERP编码(来自ProductBom.ErpCode)', 'after' => 'ttpos_item_name'])->update();
            }
        }
      
        // 2. 处理 takeout_order_item_modifier 表
        if ($this->hasTable('takeout_order_item_modifier')) {
            $modifierTable = $this->table('takeout_order_item_modifier');
            if (!$modifierTable->hasColumn('ttpos_modifier_erp_code')) {
                $modifierTable->addColumn('ttpos_modifier_erp_code', 'string', ['limit' => 50, 'default' => '', 'comment' => 'TTPOS修饰符ERP编码(来自ProductBom.ErpCode或ProductSauce.ErpCode)', 'after' => 'ttpos_modifier_name'])->update();
            }
        }
    }
}

