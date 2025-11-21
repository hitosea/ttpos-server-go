<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddOrderSourceNationalitySwitchesToCompanySetting extends Migrator
{
    const TARGET = 'all';

    /**
     * 为商户配置表新增外卖来源和国籍功能开关
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('company_setting')) {
            $table = $this->table('company_setting');
            
            // 检查字段是否不存在
            if (!$table->hasColumn('enable_order_source')) {
                $table->addColumn('enable_order_source', 'integer', ['limit' => 3, 'default' => 0, 'null' => false, 'comment' => '是否启用外卖来源：0-否；1-是', 'after' => 'is_open_marketing']);
            }
            
            if (!$table->hasColumn('enable_nationality')) {
                $table->addColumn('enable_nationality', 'integer', ['limit' => 3, 'default' => 0, 'null' => false, 'comment' => '是否启用国籍记录：0-否；1-是', 'after' => 'enable_order_source']);
            }
            
            $table->update();
        }
    }
}

