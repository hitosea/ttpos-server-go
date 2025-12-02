<?php

use think\migration\Migrator;
use think\migration\db\Column;

/**
 * 为统计销售表新增国籍字段
 */
class AddNationalityUuidToStatisticsSale extends Migrator
{
    public function change()
    {
        $table = $this->table('statistics_sale');

        $needUpdate = false;

        if (!$table->hasColumn('nationality_uuid')) {
            $table->addColumn('nationality_uuid', 'biginteger', [
                'signed'   => false,
                'null'     => false,
                'default'  => 0,
                'comment'  => '国籍UUID（0=未记录）',
                'after'    => 'order_source_uuid',
            ]);
            $needUpdate = true;
        }

        if (!$this->hasIndex('statistics_sale', 'idx_nationality_uuid')) {
            $table->addIndex(['nationality_uuid'], ['name' => 'idx_nationality_uuid']);
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

