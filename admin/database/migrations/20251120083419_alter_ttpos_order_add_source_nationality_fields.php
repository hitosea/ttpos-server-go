<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AlterTtposOrderAddSourceNationalityFields extends Migrator
{
    /**
     * 为销售账单表新增订单来源和国籍字段
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('sale_bill')) {
            $table = $this->table('sale_bill');
            
            // 检查字段是否不存在
            if (!$table->hasColumn('order_source_uuid')) {
                $table->addColumn('order_source_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'null' => false, 'comment' => '订单来源UUID（0=店内，>0=外卖）', 'after' => 'dining_method']);
            }
            
            if (!$table->hasColumn('nationality_uuid')) {
                $table->addColumn('nationality_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'null' => false, 'comment' => '国籍UUID（0=未记录）', 'after' => 'order_source_uuid']);
            }
            
            // 添加索引（检查是否已存在）
            if (!$this->hasIndex('sale_bill', 'idx_order_source_uuid')) {
                $table->addIndex(['order_source_uuid'], ['name' => 'idx_order_source_uuid']);
            }
            
            if (!$this->hasIndex('sale_bill', 'idx_nationality_uuid')) {
                $table->addIndex(['nationality_uuid'], ['name' => 'idx_nationality_uuid']);
            }
            
            $table->update();
        }
    }

    /**
     * 检查索引是否存在
     */
    private function hasIndex($tableName, $indexName)
    {
        $rows = $this->fetchAll("SHOW INDEX FROM ttpos_{$tableName} WHERE Key_name = '{$indexName}'");
        return !empty($rows);
    }
}

