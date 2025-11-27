<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddSourceToSaleBill extends Migrator
{
    /**
     * 为销售账单表新增 source 字段
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('sale_bill')) {
            $table = $this->table('sale_bill');
            
            // 检查字段是否不存在
            if (!$table->hasColumn('source')) {
                $table->addColumn('source', 'integer', [
                    'signed' => false,
                    'default' => 0,
                    'null' => false,
                    'comment' => '订单来源：0-默认值、1-收银机、2-点餐助手、3-平板、4-H5',
                    'after' => 'nationality_uuid',
                ]);
            }
            
            // 添加索引（检查是否已存在）
            if (!$this->hasIndex('sale_bill', 'idx_source')) {
                $table->addIndex(['source'], ['name' => 'idx_source']);
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

