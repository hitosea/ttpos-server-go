<?php

use think\migration\Migrator;
use think\migration\db\Column;

/**
 * 为统计销售表新增 source 字段
 */
class AddSourceToStatisticsSale extends Migrator
{
    public function change()
    {
        $table = $this->table('statistics_sale');

        $needUpdate = false;

        if (!$table->hasColumn('source')) {
            $table->addColumn('source', 'integer', [
                'signed'   => false,
                'null'     => false,
                'default'  => 0,
                'comment'  => '订单来源：0-默认值、1-收银机、2-点餐助手、3-平板、4-H5',
                'after'    => 'nationality_uuid',
            ]);
            $needUpdate = true;
        }

        if (!$this->hasIndex('statistics_sale', 'idx_source')) {
            $table->addIndex(['source'], ['name' => 'idx_source']);
            $needUpdate = true;
        }

        if ($needUpdate) {
            $table->update();
        }
    }

    /**
     * 判断索引是否存在
     */
    private function hasIndex($tableName, $indexName)
    {
        $rows = $this->fetchAll("SHOW INDEX FROM ttpos_{$tableName} WHERE Key_name = '{$indexName}'");
        return !empty($rows);
    }
}

