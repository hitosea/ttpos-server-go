<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddPurchaseTypeToPurchaseOrderTable extends Migrator
{
    /**
     * 添加 purchase_type 字段到 ttpos_purchase_order 表
     */
    public function change()
    {
        $table = $this->table('purchase_order');

        // 检查 purchase_type 字段是否已存在
        if (!$table->hasColumn('purchase_type')) {
            $table->addColumn('purchase_type', 'integer', ['default' => 1, 'comment' => '采购类型 1-外部采购 2-内部采购', 'after' => 'final_receive_time'])->update();
        }

        // 检查 warehouse_erp_code 字段是否已存在
        if (!$table->hasColumn('warehouse_erp_code')) {
            $table->addColumn('warehouse_erp_code', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '仓库ERP编码', 'after' => 'purchase_type'])->update();
        }

        // 检查 supplier_erp_code 字段是否已存在
        if (!$table->hasColumn('supplier_erp_code')) {
            $table->addColumn('supplier_erp_code', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '供应商编码', 'after' => 'supplier_name'])->update();
        }

        if (!$table->hasColumn('headquarter_status')) {
            $table->addColumn('headquarter_status', 'integer', ['null' => false, 'default' => 0, 'comment' => '总部状态：0-待提交 1-待审核 2-已通过 3-已驳回 4-部分收货 5-全部收货', 'after' => 'warehouse_erp_code'])->update();
        }

        // 检查 company_uuid 字段是否已存在
        if (!$table->hasColumn('company_uuid')) {
            $table->addColumn('company_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '公司UUID-用于识别子商户', 'after' => 'uuid'])->update();
        }

        // 检查 company_name 字段是否已存在
        if (!$table->hasColumn('company_name')) {
            $table->addColumn('company_name', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '公司名称', 'after' => 'company_uuid'])->update();
        }

        // 检查字段是否已存在
        if (!$table->hasColumn('sub_uuid')) {
            $table->addColumn('sub_uuid', 'biginteger', ['limit' => 20, 'default' => 0, 'comment' => '子订单UUID', 'after' => 'uuid'])
                  ->update();
        }
    }
}
