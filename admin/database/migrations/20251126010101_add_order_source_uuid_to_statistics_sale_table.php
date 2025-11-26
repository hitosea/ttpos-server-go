<?php

use think\migration\Migrator;
use think\migration\db\Column;

/**
 * 为统计销售表新增订单来源字段
 */
class AddOrderSourceUuidToStatisticsSaleTable extends Migrator
{
    public function change()
    {
        $table = $this->table('statistics_sale');

        $needUpdate = false;

        if (!$table->hasColumn('order_source_uuid')) {
            $table->addColumn('order_source_uuid', 'biginteger', [
                'signed'   => false,
                'null'     => false,
                'default'  => 0,
                'comment'  => '订单来源UUID（0=店内，>0=外卖/渠道）',
                'after'    => 'is_takeout',
            ]);
            $needUpdate = true;
        }

        if (!$this->hasIndex('statistics_sale', 'idx_order_source_uuid')) {
            $table->addIndex(['order_source_uuid'], ['name' => 'idx_order_source_uuid']);
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


