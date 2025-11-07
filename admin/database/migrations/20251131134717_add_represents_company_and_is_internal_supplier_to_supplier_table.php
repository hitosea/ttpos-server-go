<?php

use think\migration\Migrator;

class AddRepresentsCompanyAndIsInternalSupplierToSupplierTable extends Migrator
{
    /**
     * 添加字段到供应商表
     */
    public function change()
    {
        $table = $this->table('supplier');
        
        // 检查字段是否存在，不存在则添加
        if (!$table->hasColumn('represents_company')) {
            $table->addColumn('represents_company', 'string', ['limit' => 255, 'default' => '', 'comment' => '代表公司', 'after' => 'erp_code'])
                  ->update();
        }
        
        if (!$table->hasColumn('is_internal_supplier')) {
            $table->addColumn('is_internal_supplier', 'integer', ['limit' => 3, 'default' => 0, 'comment' => '是否内部供应商：0-否；1-是', 'after' => 'represents_company'])
                  ->update();
        }
    }
}

